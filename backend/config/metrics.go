package config

// Metrics specifies the metrics pipeline configuration
type Metrics struct {

	// Disabled specifies if the metrics provider is disabled
	Disabled *bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// Provider specifies the metrics provider: prometeus|inmem
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
}

// GetDisabled specifies if the metrics provider is disabled
func (c *Metrics) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}
