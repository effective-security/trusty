package model

import (
	"time"

	pb "github.com/ekspand/trusty/api/v1/trustypb"
)

// Certificate provides X509 Cert information
type Certificate struct {
	ID               int64     `db:"id"`
	OrgID            int64     `db:"org_id"`
	SKID             string    `db:"skid"`
	IKID             string    `db:"ikid"`
	SN               string    `db:"sn"`
	NotBefore        time.Time `db:"notbefore"`
	NotAfter         time.Time `db:"notafter"`
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
		ID:         r.ID,
		OrgID:      r.OrgID,
		SKID:       r.SKID,
		IKID:       r.IKID,
		SN:         r.SN,
		NotBefore:  r.NotBefore.Unix(),
		NotAfter:   r.NotAfter.Unix(),
		Subject:    r.Subject,
		Issuer:     r.Issuer,
		Sha256:     r.ThumbprintSha256,
		Profile:    r.Profile,
		Pem:        r.Pem,
		IssuersPem: r.IssuersPem,
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
