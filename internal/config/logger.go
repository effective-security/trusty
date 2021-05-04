package config

// Logger contains information about the configuration of a logger/log rotation
type Logger struct {

	// Directory contains where to store the log files; if value is empty, them stderr is used for output
	Directory string `json:"directory,omitempty" yaml:"directory,omitempty"`

	// MaxAgeDays controls how old files are before deletion
	MaxAgeDays int `json:"max_age_days,omitempty" yaml:"max_age_days,omitempty"`

	// MaxSizeMb contols how large a single log file can be before its rotated
	MaxSizeMb int `json:"max_size_mb,omitempty" yaml:"max_size_mb,omitempty"`
}

// RepoLogLevel contains information about the log level per repo. Use * to set up global level.
type RepoLogLevel struct {

	// Repo specifies the repo name, or '*' for all repos [Global]
	Repo string `json:"repo,omitempty" yaml:"repo,omitempty"`

	// Package specifies the package name
	Package string `json:"package,omitempty" yaml:"package,omitempty"`

	// Level specifies the log level for the repo [ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE].
	Level string `json:"level,omitempty" yaml:"level,omitempty"`
}
