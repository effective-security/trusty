package model

import "time"

// Certificate provides X509 Cert information
type Certificate struct {
	ID               int64     `db:"id"`
	OrgID            int64     `db:"org_id"`
	Skid             string    `db:"skid"`
	Ikid             string    `db:"ikid"`
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
