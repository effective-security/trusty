package logr

import (
	"github.com/effective-security/xlog"
	"github.com/go-logr/logr"
)

type prov struct {
	logger xlog.KeyValueLogger
}

// New returns logr.Logger
func New(logger xlog.KeyValueLogger) logr.Logger {
	return logr.New(&prov{logger: logger})
}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (p *prov) Init(info logr.RuntimeInfo) {}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (p *prov) Enabled(level int) bool {
	return true
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (p *prov) Info(level int, msg string, keysAndValues ...interface{}) {
	kv := append(keysAndValues, "msg", msg)
	p.logger.KV(xlog.INFO, kv...)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (p *prov) Error(err error, msg string, keysAndValues ...interface{}) {
	kv := append(keysAndValues, "msg", msg, "err", err.Error())
	p.logger.KV(xlog.ERROR, kv...)
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (p *prov) WithValues(keysAndValues ...interface{}) logr.LogSink {
	p.logger = p.logger.WithValues(keysAndValues...)
	return p
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (p *prov) WithName(name string) logr.LogSink {
	p.logger = p.logger.WithValues("src", name)
	return p
}
