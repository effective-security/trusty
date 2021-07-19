package httpclient

import (
	"crypto/tls"
)

// Config of the client, particularly around error & retry handling
type Config struct {

	// TLS contains all the TLS/HTTPS configuration, at a minimum Certificates should be set to the client cert
	// [a client cert is required to communicate with Keydi]
	TLS *tls.Config `json:"-"`
}

var defaultConfig = Config{
	TLS: &tls.Config{
		MinVersion: tls.VersionTLS12,
	},
}

// NewConfig returns a new Config populated with recommended defaults
func NewConfig() *Config {
	c := new(Config)
	defaultConfig.copyTo(c)
	return c
}

func (c *Config) copyTo(dst *Config) {
	if c.TLS != nil {
		dst.TLS = c.TLS.Clone()
	}
}
