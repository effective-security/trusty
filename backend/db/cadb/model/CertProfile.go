package model

import (
	"time"
)

// CertProfile provides Cert Profile configuration
type CertProfile struct {
	ID          uint64    `db:"id"`
	Label       string    `db:"label"`
	IssuerLabel string    `db:"issuer_label"`
	Config      string    `db:"config"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
