package model

import (
	"time"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RootCertificate provides X509 Root Certificate information
type RootCertificate struct {
	ID               int64     `db:"id"`
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

// ToDTO returns DTO
func (list RootCertificates) ToDTO() []*pb.RootCertificate {
	dto := make([]*pb.RootCertificate, len(list))
	for i, r := range list {
		dto[i] = r.ToDTO()
	}
	return dto
}

// Find a cert by ID
func (list RootCertificates) Find(id int64) *RootCertificate {
	for _, m := range list {
		if m.ID == id {
			return m
		}
	}
	return nil
}
