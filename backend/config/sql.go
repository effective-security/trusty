package config

// SQL specifies the configuration for SQL provider.
type SQL struct {
	// DataSource specifies the connection string. It can be prefixed with file:// or env:// to load the source from a file or environment variable.
	DataSource string `json:"data_source,omitempty" yaml:"data_source,omitempty"`

	// MigrationsDir specifies the directory that contains migrations.
	MigrationsDir string `json:"migrations_dir,omitempty" yaml:"migrations_dir,omitempty"`

	// ForceVersion specifies to force version.
	ForceVersion int `json:"force_version,omitempty" yaml:"force_version,omitempty"`

	// MigrateVersion specifies to migrate version.
	MigrateVersion int `json:"migrate_version,omitempty" yaml:"migrate_version,omitempty"`
}
