package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mylxsw/redis-tui/api"
	"github.com/mylxsw/redis-tui/config"
	"github.com/mylxsw/redis-tui/core"
	"github.com/mylxsw/redis-tui/tui"
)

var conf = config.Config{}

var Version string
var GitCommit string

// parseRedisURL parses a Redis URL and updates the config
// Format: redis://[user[:password]@]host[:port][/db][?option=value]
// or: rediss:// for TLS connections
func parseRedisURL(redisURL string, conf *config.Config) error {
	u, err := url.Parse(redisURL)
	if err != nil {
		return err
	}

	// Check scheme
	switch u.Scheme {
	case "redis":
		conf.TLS = false
	case "rediss":
		conf.TLS = true
	default:
		return fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	// Parse host and port
	if u.Hostname() != "" {
		conf.Host = u.Hostname()
	}
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return fmt.Errorf("invalid port: %s", u.Port())
		}
		conf.Port = port
	}

	// Parse password
	if password, ok := u.User.Password(); ok {
		conf.Password = password
	}

	// Parse database from path
	if u.Path != "" && u.Path != "/" {
		dbStr := strings.TrimPrefix(u.Path, "/")
		db, err := strconv.Atoi(dbStr)
		if err != nil {
			return fmt.Errorf("invalid database number: %s", dbStr)
		}
		conf.DB = db
	}

	// Parse query parameters
	query := u.Query()
	if cluster := query.Get("cluster"); cluster == "true" {
		conf.Cluster = true
	}

	return nil
}

func main() {
	// Initialize defaults
	conf.Host = "127.0.0.1"
	conf.Port = 6379
	conf.TLSVerify = true

	// Set defaults from REDIS_URL if present
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if err := parseRedisURL(redisURL, &conf); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing REDIS_URL: %v\n", err)
			os.Exit(1)
		}
	}

	flag.StringVar(&conf.Host, "h", conf.Host, "Server hostname")
	flag.IntVar(&conf.Port, "p", conf.Port, "Server port")
	flag.StringVar(&conf.Password, "a", conf.Password, "Password to use when connecting to the server")
	flag.IntVar(&conf.DB, "n", conf.DB, "Database number")
	flag.BoolVar(&conf.Cluster, "c", conf.Cluster, "Enable cluster mode")
	flag.BoolVar(&conf.Debug, "vvv", conf.Debug, "Enable debug mode")
	flag.BoolVar(&conf.TLS, "tls", conf.TLS, "Enable TLS/SSL connection")
	flag.StringVar(&conf.TLSCert, "tls-cert", conf.TLSCert, "TLS client certificate file")
	flag.StringVar(&conf.TLSKey, "tls-key", conf.TLSKey, "TLS client key file")
	flag.StringVar(&conf.TLSCACert, "tls-ca-cert", conf.TLSCACert, "TLS CA certificate file")
	flag.BoolVar(&conf.TLSVerify, "tls-verify", conf.TLSVerify, "Enable TLS certificate verification")

	var showVersion bool
	var redisURLFlag string
	flag.BoolVar(&showVersion, "v", false, "Show version and exit")
	flag.StringVar(&redisURLFlag, "url", "", "Redis URL (overrides other connection options)")

	flag.Parse()

	// If -url flag is provided, it overrides everything including env var
	if redisURLFlag != "" {
		if err := parseRedisURL(redisURLFlag, &conf); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing Redis URL: %v\n", err)
			os.Exit(1)
		}
	}

	if len(GitCommit) > 8 {
		GitCommit = GitCommit[:8]
	}

	if showVersion {
		fmt.Printf("Version: %s\nGitCommit: %s\n", Version, GitCommit)

		return
	}

	outputChan := make(chan core.OutputMessage, 100)
	if err := tui.NewRedisTUI(api.NewRedisClient(conf, outputChan), 100, Version, GitCommit, outputChan, conf).Start(); err != nil {
		panic(err)
	}
}
