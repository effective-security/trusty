package model

import (
	"crypto/x509"
	"encoding/base64"
	"math/big"
	"time"

	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Certificate provides X509 Cert information
type Certificate struct {
	ID               uint64    `db:"id"`
	OrgID            uint64    `db:"org_id"`
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
	Locations        []string  `db:"locations"`
}

// Certificates defines a list of Certificate
type Certificates []*Certificate

// ToPB returns protobuf
func (r *Certificate) ToPB() *pb.Certificate {
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
		Locations:    r.Locations,
	}
}

// FileName returns  a generated file name for publisher,
// in {IKID[:4] / Base64(sn[:9])} format
func (r *Certificate) FileName() string {
	sn := r.SerialNumber[:12]
	n := new(big.Int)
	n, ok := n.SetString(r.SerialNumber, 10)
	if ok {
		b := n.Bytes()
		l := len(b)
		if l > 9 {
			l = 9
		}
		sn = base64.RawURLEncoding.EncodeToString(b[:l])
	}
	return r.IKID[:4] + "/" + sn
}

/*
// ToDTO returns ToDTO
func (r *Certificate) ToDTO() *v1.Certificate {
	return &v1.Certificate{
		ID:           strconv.FormatUint(r.ID, 10),
		OrgID:        strconv.FormatUint(r.OrgID, 10),
		SKID:         r.SKID,
		IKID:         r.IKID,
		SerialNumber: r.SerialNumber,
		NotBefore:    r.NotBefore.UTC(),
		NotAfter:     r.NotAfter.UTC(),
		Subject:      r.Subject,
		Issuer:       r.Issuer,
		Sha256:       r.ThumbprintSha256,
		Profile:      r.Profile,
		Pem:          r.Pem,
		IssuersPem:   r.IssuersPem,
		Locations:    r.Locations,
	}
}
*/

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
		Locations:        r.Locations,
	}
}

// NewCertificate returns Certificate
func NewCertificate(r *x509.Certificate, orgID uint64, profile, pem, issuersPem string, locations []string) *Certificate {
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
		Locations:        locations,
	}
}

// ToDTO returns DTO
func (list Certificates) ToDTO() []*pb.Certificate {
	dto := make([]*pb.Certificate, len(list))
	for i, r := range list {
		dto[i] = r.ToPB()
	}
	return dto
}

// Find a cert by ID
func (list Certificates) Find(id uint64) *Certificate {
	for _, m := range list {
		if m.ID == id {
			return m
		}
	}
	return nil
}
