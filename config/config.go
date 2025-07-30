package config

type Config struct {
	Host          string
	Port          int
	Password      string
	DB            int
	Cluster       bool
	Debug         bool
	TLS           bool
	TLSCert       string
	TLSKey        string
	TLSCACert          string
	TLSVerify          bool
}
