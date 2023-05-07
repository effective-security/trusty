package model

import (
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/trusty/api/v1/pb"
)

// Crl provides X509 CRL information
type Crl struct {
	ID         uint64   `db:"id"`
	IKID       string   `db:"ikid"`
	ThisUpdate xdb.Time `db:"this_update"`
	NextUpdate xdb.Time `db:"next_update"`
	Issuer     string   `db:"issuer"`
	Pem        string   `db:"pem"`
}

// ToDTO returns DTO
func (r *Crl) ToDTO() *pb.Crl {
	return &pb.Crl{
		ID:         r.ID,
		IKID:       r.IKID,
		ThisUpdate: r.ThisUpdate.String(),
		NextUpdate: r.NextUpdate.String(),
		Issuer:     r.Issuer,
		Pem:        r.Pem,
	}
}
