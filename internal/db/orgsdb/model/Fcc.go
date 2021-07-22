package model

import (
	"time"
)

// FccFRNResponse represents a cached FRN response.
type FccFRNResponse struct {
	FilerID   uint64    `db:"filer_id"`
	UpdatedAt time.Time `db:"updated_at"`
	Response  string    `db:"json"`
}
