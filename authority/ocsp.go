package authority

import (
	"crypto"
	"crypto/x509/pkix"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ocsp"
)

// revocationReasonCodes is a map between string reason codes
// to integers as defined in RFC 5280
var revocationReasonCodes = map[string]int{
	"unspecified":          ocsp.Unspecified,
	"keycompromise":        ocsp.KeyCompromise,
	"cacompromise":         ocsp.CACompromise,
	"affiliationchanged":   ocsp.AffiliationChanged,
	"superseded":           ocsp.Superseded,
	"cessationofoperation": ocsp.CessationOfOperation,
	"certificatehold":      ocsp.CertificateHold,
	"removefromcrl":        ocsp.RemoveFromCRL,
	"privilegewithdrawn":   ocsp.PrivilegeWithdrawn,
	"aacompromise":         ocsp.AACompromise,
}

const (
	// OCSPStatusGood specifies name for good status
	OCSPStatusGood = "good"
	// OCSPStatusRevoked specifies name for revoked status
	OCSPStatusRevoked = "revoked"
	// OCSPStatusUnknown specifies name for unknown status
	OCSPStatusUnknown = "unknown"
)

// OCSPStatusCode is a map between string statuses sent by cli/api
// to ocsp int statuses
var OCSPStatusCode = map[string]int{
	OCSPStatusGood:    ocsp.Good,
	OCSPStatusRevoked: ocsp.Revoked,
	OCSPStatusUnknown: ocsp.Unknown,
}

// OCSPSignRequest represents the desired contents of a
// specific OCSP response.
type OCSPSignRequest struct {
	SerialNumber *big.Int
	Status       string
	Reason       int
	RevokedAt    time.Time
	Extensions   []pkix.Extension
	// IssuerHash is the hashing function used to hash the issuer subject and public key
	// in the OCSP response. Valid values are crypto.SHA1, crypto.SHA256, crypto.SHA384,
	// and crypto.SHA512. If zero, the default is crypto.SHA1.
	IssuerHash crypto.Hash
	// If provided ThisUpdate will override the default usage of time.Now().Truncate(time.Hour)
	ThisUpdate *time.Time
	// If provided NextUpdate will override the default usage of ThisUpdate.Add(signerInterval)
	NextUpdate *time.Time
}

// OCSPReasonStringToCode tries to convert a reason string to an integer code
func OCSPReasonStringToCode(reason string) (reasonCode int, err error) {
	// default to 0
	if reason == "" {
		return 0, nil
	}

	reasonCode, present := revocationReasonCodes[strings.ToLower(reason)]
	if !present {
		reasonCode, err = strconv.Atoi(reason)
		if err != nil {
			return
		}
		if reasonCode > ocsp.AACompromise || reasonCode < ocsp.Unspecified {
			return 0, errors.Errorf("invalid status: %s", reason)
		}
	}

	return
}

// SignOCSP return an OCSP response.
func (i *Issuer) SignOCSP(req *OCSPSignRequest) ([]byte, error) {
	var thisUpdate, nextUpdate time.Time
	if req.ThisUpdate != nil {
		thisUpdate = *req.ThisUpdate
	} else {
		// Round thisUpdate times down to the nearest minute
		thisUpdate = time.Now().Truncate(time.Minute)
	}
	if req.NextUpdate != nil {
		nextUpdate = *req.NextUpdate
	} else {
		nextUpdate = thisUpdate.Add(i.ocspExpiry)
	}

	status, ok := OCSPStatusCode[req.Status]
	if !ok {
		return nil, errors.Errorf("invalid status: %s", req.Status)
	}

	template := ocsp.Response{
		Status:       status,
		SerialNumber: req.SerialNumber,
		ThisUpdate:   thisUpdate.UTC(),
		NextUpdate:   nextUpdate.UTC(),
		//Certificate:     certificate, // if responder cert is issuer, then no need to include
		ExtraExtensions: req.Extensions,
		IssuerHash:      req.IssuerHash,
	}

	if status == ocsp.Revoked {
		template.RevokedAt = req.RevokedAt
		template.RevocationReason = req.Reason
	}

	res, err := ocsp.CreateResponse(i.bundle.Cert, i.bundle.Cert, template, i.signer)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
