package model

import "time"

// RootCertificate provides X509 Root Certificate information
type RootCertificate struct {
	ID               int64     `db:"id"`
	OwnerID          int64     `db:"owner_id"`
	Skid             string    `db:"skid"`
	NotBefore        time.Time `db:"notbefore"`
	NotAfter         time.Time `db:"notafter"`
	Subject          string    `db:"subject"`
	ThumbprintSha256 string    `db:"sha256"`
	Trust            int       `db:"trust"`
	Pem              string    `db:"pem"`
}
