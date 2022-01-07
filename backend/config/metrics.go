package config

// Metrics specifies the metrics pipeline configuration
type Metrics struct {
	// Disabled specifies if the metrics provider is disabled
	Disabled *bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// Provider specifies the metrics provider: prometheus|inmem
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// Prometheus provider config
	Prometheus *Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}

// GetDisabled specifies if the metrics provider is disabled
func (c *Metrics) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// Prometheus provider config
type Prometheus struct {
	Addr string `json:"addr,omitempty" yaml:"addr,omitempty"`
}
