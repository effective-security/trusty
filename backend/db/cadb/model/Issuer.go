package model

import (
	"time"
)

// Issuer provides Issuer configuration
type Issuer struct {
	ID        uint64    `db:"id"`
	Label     string    `db:"label"`
	Config    string    `db:"config"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
