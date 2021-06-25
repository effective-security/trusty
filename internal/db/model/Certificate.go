package model

import (
	"crypto/x509"
	"time"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/go-phorce/dolly/xpki/certutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Certificate provides X509 Cert information
type Certificate struct {
	ID               int64     `db:"id"`
	OrgID            int64     `db:"org_id"`
	SKID             string    `db:"skid"`
	IKID             string    `db:"ikid"`
	SerialNumber     string    `db:"serial_number"`
	NotBefore        time.Time `db:"not_before"`
	NotAfter         time.Time `db:"no_tafter"`
	Subject          string    `db:"subject"`
	Issuer           string    `db:"issuer"`
	ThumbprintSha256 string    `db:"sha256"`
	Profile          string    `db:"profile"`
	Pem              string    `db:"pem"`
	IssuersPem       string    `db:"issuers_pem"`
}

// Certificates defines a list of Certificate
type Certificates []*Certificate

// ToDTO returns DTO
func (r *Certificate) ToDTO() *pb.Certificate {
	return &pb.Certificate{
		Id:           r.ID,
		OrgId:        r.OrgID,
		Skid:         r.SKID,
		Ikid:         r.IKID,
		SerialNumber: r.SerialNumber,
		NotBefore:    timestamppb.New(r.NotBefore),
		NotAfter:     timestamppb.New(r.NotAfter),
		Subject:      r.Subject,
		Issuer:       r.Issuer,
		Sha256:       r.ThumbprintSha256,
		Profile:      r.Profile,
		Pem:          r.Pem,
		IssuersPem:   r.IssuersPem,
	}
}

// CertificateFromPB returns Certificate
func CertificateFromPB(r *pb.Certificate) *Certificate {
	return &Certificate{
		ID:               r.Id,
		OrgID:            r.OrgId,
		SKID:             r.Skid,
		IKID:             r.Ikid,
		SerialNumber:     r.SerialNumber,
		NotBefore:        r.NotBefore.AsTime().UTC(),
		NotAfter:         r.NotAfter.AsTime().UTC(),
		Subject:          r.Subject,
		Issuer:           r.Issuer,
		ThumbprintSha256: r.Sha256,
		Profile:          r.Profile,
		Pem:              r.Pem,
		IssuersPem:       r.IssuersPem,
	}
}

// NewCertificate returns Certificate
func NewCertificate(r *x509.Certificate, orgID int64, profile, pem, issuersPem string) *Certificate {
	return &Certificate{
		//ID:               r.Id,
		OrgID:            orgID,
		SKID:             certutil.GetSubjectKeyID(r),
		IKID:             certutil.GetAuthorityKeyID(r),
		SerialNumber:     r.SerialNumber.String(),
		NotBefore:        r.NotBefore.UTC(),
		NotAfter:         r.NotAfter.UTC(),
		Subject:          r.Subject.String(),
		Issuer:           r.Issuer.String(),
		ThumbprintSha256: certutil.SHA256Hex(r.Raw),
		Profile:          profile,
		Pem:              pem,
		IssuersPem:       issuersPem,
	}
}

// ToDTO returns DTO
func (list Certificates) ToDTO() []*pb.Certificate {
	dto := make([]*pb.Certificate, len(list))
	for i, r := range list {
		dto[i] = r.ToDTO()
	}
	return dto
}

// Find a cert by ID
func (list Certificates) Find(id int64) *Certificate {
	for _, m := range list {
		if m.ID == id {
			return m
		}
	}
	return nil
}
