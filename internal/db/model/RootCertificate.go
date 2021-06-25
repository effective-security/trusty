package model

import (
	"crypto/x509"
	"time"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/go-phorce/dolly/xpki/certutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RootCertificate provides X509 Root Certificate information
type RootCertificate struct {
	ID               uint64    `db:"id"`
	SKID             string    `db:"skid"`
	NotBefore        time.Time `db:"not_before"`
	NotAfter         time.Time `db:"no_tafter"`
	Subject          string    `db:"subject"`
	ThumbprintSha256 string    `db:"sha256"`
	Trust            int       `db:"trust"`
	Pem              string    `db:"pem"`
}

// RootCertificates defines a list of RootCertificate
type RootCertificates []*RootCertificate

// ToDTO returns DTO
func (r *RootCertificate) ToDTO() *pb.RootCertificate {
	return &pb.RootCertificate{
		Id:        r.ID,
		Skid:      r.SKID,
		NotBefore: timestamppb.New(r.NotBefore),
		NotAfter:  timestamppb.New(r.NotAfter),
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
		NotBefore:        r.NotBefore.UTC(),
		NotAfter:         r.NotAfter.UTC(),
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
