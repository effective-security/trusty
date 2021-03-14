package model

import "time"

// RevokedCertificate provides X509 Cert information
type RevokedCertificate struct {
	Certificate Certificate
	RevokedAt   time.Time `db:"revoked_at"`
	Reason      int       `db:"reason"`
}
