package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/mylxsw/redis-tui/config"
	"github.com/mylxsw/redis-tui/core"

	"github.com/gdamore/tcell"
	"github.com/go-redis/redis"
)

type RedisClient interface {
	Keys(pattern string) *redis.StringSliceCmd
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
	Type(key string) *redis.StatusCmd
	TTL(key string) *redis.DurationCmd
	Get(key string) *redis.StringCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	SMembers(key string) *redis.StringSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	HKeys(key string) *redis.StringSliceCmd
	HGet(key, field string) *redis.StringCmd
	Process(cmd redis.Cmder) error
	Do(args ...interface{}) *redis.Cmd
	Info(section ...string) *redis.StringCmd
	Subscribe(channels ...string) *redis.PubSub
	ConfigSet(parameter, value string) *redis.StatusCmd
}

// createTLSConfig creates a TLS configuration based on the provided settings
func createTLSConfig(conf config.Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !conf.TLSVerify,
	}

	// Load client certificate if provided
	if conf.TLSCert != "" && conf.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(conf.TLSCert, conf.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if conf.TLSCACert != "" {
		caCert, err := ioutil.ReadFile(conf.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// NewRedisClient create a new redis client which wraps single or cluster client
func NewRedisClient(conf config.Config, outputChan chan core.OutputMessage) RedisClient {
	var tlsConfig *tls.Config
	var err error

	// Create TLS configuration if enabled
	if conf.TLS {
		tlsConfig, err = createTLSConfig(conf)
		if err != nil {
			// Log error to output channel
			outputChan <- core.OutputMessage{
				Color:   tcell.ColorRed,
				Message: fmt.Sprintf("TLS configuration error: %v", err),
			}
			// Continue without TLS rather than failing completely
			tlsConfig = nil
		}
	}

	if conf.Cluster {
		options := &redis.ClusterOptions{
			Addrs:     []string{fmt.Sprintf("%s:%d", conf.Host, conf.Port)},
			Password:  conf.Password,
			TLSConfig: tlsConfig,
		}

		return redis.NewClusterClient(options)
	}

	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		DB:           conf.DB,
		Password:     conf.Password,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  2 * time.Second,
		TLSConfig:    tlsConfig,
	}

	client := redis.NewClient(options)
	if conf.Debug {
		client.WrapProcess(func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
			return func(cmd redis.Cmder) error {

				outputChan <- core.OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("redis: <%s>", cmd)}
				err := oldProcess(cmd)

				return err
			}
		})
	}

	return client
}

func RedisExecute(client RedisClient, command string) (interface{}, error) {
	stringArgs := strings.Split(command, " ")
	var args = make([]interface{}, len(stringArgs))
	for i, s := range stringArgs {
		args[i] = s
	}

	return client.Do(args...).Result()
}

var redisKeys = make([]string, 0)
var redisKeysLastUpdate time.Time
var redisLock sync.RWMutex

func RedisKeys(client RedisClient, pattern string) ([]string, error) {
	keys, err := KeysWithLimit(client, pattern, -1)
	if err != nil {
		return nil, nil
	}

	return keys, nil
}

func RedisAllKeys(client RedisClient, cache bool) ([]string, error) {
	redisLock.RLock()
	if cache && redisKeysLastUpdate.After(time.Now().Add(60*time.Second)) {
		redisLock.RUnlock()
		return redisKeys, nil
	}
	redisLock.RUnlock()

	redisLock.Lock()
	defer redisLock.Unlock()

	keys, err := KeysWithLimit(client, "*", 10)
	if err != nil {
		return nil, err
	}

	redisKeys = keys
	redisKeysLastUpdate = time.Now()

	return keys, nil
}

func KeysWithLimit(client RedisClient, key string, maxScanCount int) (redisKeys []string, err error) {
	var cursor uint64 = 0
	var keys []string

	var scanCount = 0
	for scanCount < maxScanCount || maxScanCount == -1 {
		scanCount++

		keys, cursor, err = client.Scan(cursor, key, 100).Result()
		if err != nil {
			return
		}

		redisKeys = append(redisKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	return
}

func RedisServerInfo(conf config.Config, client RedisClient) (string, error) {
	res, err := client.Info().Result()
	if err != nil {
		return "", err
	}

	var kvpairs = make(map[string]string)
	for _, kv := range strings.Split(res, "\n") {
		if strings.HasPrefix(kv, "#") || kv == "" {
			continue
		}

		pair := strings.SplitN(kv, ":", 2)
		if len(pair) != 2 {
			continue
		}

		kvpairs[pair[0]] = pair[1]
	}

	keySpace := "-"
	if ks, ok := kvpairs[fmt.Sprintf("db%d", conf.DB)]; ok {
		keySpace = ks
	}
	return fmt.Sprintf(" RedisVersion: %s    Memory: %s    Server: %s:%d/%d\n KeySpace: %s", kvpairs["redis_version"], kvpairs["used_memory_human"], conf.Host, conf.Port, conf.DB, keySpace), nil
}
