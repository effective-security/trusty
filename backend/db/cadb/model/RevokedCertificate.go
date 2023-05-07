package model

import (
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/trusty/api/v1/pb"
)

// RevokedCertificate provides X509 Cert information
type RevokedCertificate struct {
	Certificate Certificate
	RevokedAt   xdb.Time `db:"revoked_at"`
	Reason      int      `db:"reason"`
}

// ToDTO returns DTO
func (r *RevokedCertificate) ToDTO() *pb.RevokedCertificate {
	return &pb.RevokedCertificate{
		Certificate: r.Certificate.ToPB(),
		RevokedAt:   r.RevokedAt.String(),
		Reason:      pb.Reason(r.Reason),
	}
}

// RevokedCertificates defines a list of RevokedCertificate
type RevokedCertificates []*RevokedCertificate

// ToDTO returns DTO
func (list RevokedCertificates) ToDTO() []*pb.RevokedCertificate {
	dto := make([]*pb.RevokedCertificate, len(list))
	for i, r := range list {
		dto[i] = r.ToDTO()
	}
	return dto
}

// Find a cert by ID
func (list RevokedCertificates) Find(id uint64) *RevokedCertificate {
	for _, m := range list {
		if m.Certificate.ID == id {
			return m
		}
	}
	return nil
}
