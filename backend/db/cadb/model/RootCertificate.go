package model

import (
	"crypto/x509"

	pb "github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xpki/certutil"
)

// RootCertificate provides X509 Root Certificate information
type RootCertificate struct {
	ID               uint64   `db:"id"`
	SKID             string   `db:"skid"`
	NotBefore        xdb.Time `db:"not_before"`
	NotAfter         xdb.Time `db:"no_tafter"`
	Subject          string   `db:"subject"`
	ThumbprintSha256 string   `db:"sha256"`
	Trust            int      `db:"trust"`
	Pem              string   `db:"pem"`
}

// RootCertificates defines a list of RootCertificate
type RootCertificates []*RootCertificate

// ToDTO returns DTO
func (r *RootCertificate) ToDTO() *pb.RootCertificate {
	return &pb.RootCertificate{
		ID:        r.ID,
		SKID:      r.SKID,
		NotBefore: r.NotBefore.String(),
		NotAfter:  r.NotAfter.String(),
		Subject:   r.Subject,
		Sha256:    r.ThumbprintSha256,
		Trust:     pb.Trust(r.Trust),
		Pem:       r.Pem,
	}
}

// NewRootCertificate returns RootCertificate
func NewRootCertificate(r *x509.Certificate, trust int, pem string) *RootCertificate {
	return &RootCertificate{
		//ID:
		SKID:             certutil.GetSubjectKeyID(r),
		NotBefore:        xdb.Time(r.NotBefore),
		NotAfter:         xdb.Time(r.NotAfter),
		Subject:          r.Subject.String(),
		ThumbprintSha256: certutil.SHA256Hex(r.Raw),
		Trust:            trust,
		Pem:              pem,
	}
}

// ToDTO returns DTO
func (list RootCertificates) ToDTO() []*pb.RootCertificate {
	dto := make([]*pb.RootCertificate, len(list))
	for i, r := range list {
		dto[i] = r.ToDTO()
	}
	return dto
}

// Find a cert by ID
func (list RootCertificates) Find(id uint64) *RootCertificate {
	for _, m := range list {
		if m.ID == id {
			return m
		}
	}
	return nil
}
