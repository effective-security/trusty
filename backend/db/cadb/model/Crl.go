package model

import (
	"time"

	"github.com/effective-security/trusty/api/v1/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Crl provides X509 CRL information
type Crl struct {
	ID         uint64    `db:"id"`
	IKID       string    `db:"ikid"`
	ThisUpdate time.Time `db:"this_update"`
	NextUpdate time.Time `db:"next_update"`
	Issuer     string    `db:"issuer"`
	Pem        string    `db:"pem"`
}

// ToDTO returns DTO
func (r *Crl) ToDTO() *pb.Crl {
	return &pb.Crl{
		Id:         r.ID,
		Ikid:       r.IKID,
		ThisUpdate: timestamppb.New(r.ThisUpdate),
		NextUpdate: timestamppb.New(r.NextUpdate),
		Issuer:     r.Issuer,
		Pem:        r.Pem,
	}
}
