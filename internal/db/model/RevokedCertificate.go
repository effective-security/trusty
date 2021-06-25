package model

import "time"

// RevokedCertificate provides X509 Cert information
type RevokedCertificate struct {
	Certificate Certificate
	RevokedAt   time.Time `db:"revoked_at"`
	Reason      int       `db:"reason"`
}

// RevokedCertificates defines a list of RevokedCertificate
type RevokedCertificates []*RevokedCertificate

// Find a cert by ID
func (list RevokedCertificates) Find(id uint64) *RevokedCertificate {
	for _, m := range list {
		if m.Certificate.ID == id {
			return m
		}
	}
	return nil
}
