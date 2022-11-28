package config

// Metrics specifies the metrics pipeline configuration
type Metrics struct {
	// Disabled specifies if the metrics provider is disabled
	Disabled *bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// Provider specifies the metrics provider: prometheus|inmem
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// Prefix specifies the prefix added to all metrics
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Prometheus provider config
	Prometheus *Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`

	EnableRuntimeMetrics bool `json:"runtime_metrics,omitempty" yaml:"runtime_metrics,omitempty"`

	GlobalTags []string `json:"global_tags,omitempty" yaml:"global_tags,omitempty"`
}

// GetDisabled specifies if the metrics provider is disabled
func (c *Metrics) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// Prometheus provider config
type Prometheus struct {
	Addr string `json:"addr,omitempty" yaml:"addr,omitempty"`
}
