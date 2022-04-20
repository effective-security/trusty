package config

import (
	"time"

	"github.com/effective-security/porto/gserver"
)

// TrustyClient specifies configurations for the client to connect to the cluster
type TrustyClient struct {
	// ClientTLS describes the TLS certs used to connect to the cluster
	ClientTLS gserver.TLSInfo `json:"client_tls,omitempty" yaml:"client_tls,omitempty"`

	// ServerURL specifies URLs for each server
	ServerURL map[string][]string `json:"server_url,omitempty" yaml:"server_url,omitempty"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout time.Duration `json:"dial_timeout,omitempty" yaml:"dial_timeout,omitempty"`

	// DialKeepAliveTime is the time after which client pings the server to see if
	// transport is alive.
	DialKeepAliveTime time.Duration `json:"dial_keep_alive_time,omitempty" yaml:"dial_keep_alive_time,omitempty"`

	// DialKeepAliveTimeout is the time that the client waits for a response for the
	// keep-alive probe. If the response is not received in this time, the connection is closed.
	DialKeepAliveTimeout time.Duration `json:"dial_keep_alive_timeout,omitempty" yaml:"dial_keep_alive_timeout,omitempty"`
}
