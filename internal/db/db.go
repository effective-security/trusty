package db

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/juju/errors"
)

// Max values for strings
const (
	MaxLenForName     = 64
	MaxLenForEmail    = 160
	MaxLenForShortURL = 256
)

// IDGenerator defines an interface to generate unique ID accross the cluster
type IDGenerator interface {
	// NextID generates a next unique ID.
	NextID() (uint64, error)
}

// Validator provides schema validation interface
type Validator interface {
	// Validate returns error if the model is not valid
	Validate() error
}

// Validate returns error if the model is not valid
func Validate(m interface{}) error {
	if v, ok := m.(Validator); ok {
		return v.Validate()
	}
	return nil
}

// NullTime from *time.Time
func NullTime(val *time.Time) sql.NullTime {
	if val == nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: *val, Valid: true}
}

// String returns string
func String(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

// ID returns id from the string
func ID(id string) (uint64, error) {
	i64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return i64, nil
}
