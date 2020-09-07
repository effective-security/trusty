package trustyserver

import (
	"errors"
)

var (
	// ErrUnknownMethod specifies UnknownMethod error
	ErrUnknownMethod = errors.New("unknown method")
	// ErrAlreadyClosed specifies AlreadyClosed error
	ErrAlreadyClosed = errors.New("already closed")
	// ErrStopped specifies Stopped error
	ErrStopped = errors.New("server stopped")
	// ErrCanceled specifies Canceled error
	ErrCanceled = errors.New("request cancelled")
	// ErrTimeout specifies Timeout error
	ErrTimeout = errors.New("request timed out")
)
