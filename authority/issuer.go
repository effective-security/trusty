package authority

import (
	"crypto"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
)

var (
	supportedKeyHash = []crypto.Hash{crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512}
)

// Issuer of certificates
type Issuer struct {
	label          string
	skid           string // Subject Key ID
	rkid           string // Root Key ID
	signer         crypto.Signer
	bundle         *certutil.Bundle
	crlNextUpdate  time.Duration
	ocspNextUpdate time.Duration
	crlURL         string
	aiaURL         string
	ocspURL        string
	// caCerts contains PEM encoded certs for the issuer,
	// this bundle includes Issuing cert itself and its parents.
	caCerts string

	keyHash  map[crypto.Hash][]byte
	nameHash map[crypto.Hash][]byte
}

// Bundle returns certificates bundle
func (i *Issuer) Bundle() *certutil.Bundle {
	return i.bundle
}

// PEM returns PEM encoded certs for the issuer
func (i *Issuer) PEM() string {
	return i.caCerts
}

// CrlURL returns CRL DP URL
func (i *Issuer) CrlURL() string {
	return i.crlURL
}

// OcspURL returns OCSP URL
func (i *Issuer) OcspURL() string {
	return i.ocspURL
}

// AiaURL returns AIA URL
func (i *Issuer) AiaURL() string {
	return i.aiaURL
}

// Label returns label of the issuer
func (i *Issuer) Label() string {
	return i.label
}

// SubjectKID returns Subject Key ID
func (i *Issuer) SubjectKID() string {
	return i.skid
}

// RootKID returns Root Key ID
func (i *Issuer) RootKID() string {
	return i.rkid
}

// Signer returns crypto.Signer
func (i *Issuer) Signer() crypto.Signer {
	return i.signer
}

// KeyHash returns key hash
func (i *Issuer) KeyHash(h crypto.Hash) []byte {
	return i.keyHash[h]
}

// NewIssuer creates Issuer from provided configuration
func NewIssuer(cfg *config.Issuer, caCfg *Config, prov *cryptoprov.Crypto) (*Issuer, error) {
	// ensure that signer can be created before the key is generated
	cryptoSigner, err := NewSignerFromFromFile(
		prov,
		cfg.KeyFile)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to create signer")
	}

	// Build the bundle and register the CA cert
	var intCAbytes, rootBytes []byte
	if cfg.CABundleFile != "" {
		intCAbytes, err = ioutil.ReadFile(cfg.CABundleFile)
		if err != nil {
			return nil, errors.Annotate(err, "failed to load ca-bundle")
		}
	}

	if cfg.RootBundleFile != "" {
		rootBytes, err = ioutil.ReadFile(cfg.RootBundleFile)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to load root-bundle")
		}
	}

	certBytes, err := ioutil.ReadFile(cfg.CertFile)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to load cert")
	}

	bundle, status, err := certutil.VerifyBundleFromPEM(certBytes, intCAbytes, rootBytes)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create signing CA cert bundle")
	}
	if status.IsUntrusted() {
		return nil, errors.Annotatef(err, "bundle is invalid: label=%s, cn=%q, expiresAt=%q, expiringSKU=[%v], untrusted=[%v]",
			cfg.Label,
			bundle.Subject.CommonName,
			bundle.Expires.Format(time.RFC3339),
			strings.Join(status.ExpiringSKIs, ","),
			strings.Join(status.Untrusted, ","),
		)
	}

	crl := strings.Replace(caCfg.DefaultCrlURL, "${ISSUER_ID}", bundle.SubjectID, -1)
	aia := strings.Replace(caCfg.DefaultAiaURL, "${ISSUER_ID}", bundle.SubjectID, -1)
	ocsp := strings.Replace(caCfg.DefaultOcspURL, "${ISSUER_ID}", bundle.SubjectID, -1)

	keyHash := make(map[crypto.Hash][]byte)
	nameHash := make(map[crypto.Hash][]byte)

	for _, h := range supportedKeyHash {
		// OCSP requires Hash of the Key without Tag:
		/// issuerKeyHash is the hash of the issuer's public key.  The hash
		// shall be calculated over the value (excluding tag and length) of
		// the subject public key field in the issuer's certificate.
		var publicKeyInfo struct {
			Algorithm pkix.AlgorithmIdentifier
			PublicKey asn1.BitString
		}
		_, err = asn1.Unmarshal(bundle.Cert.RawSubjectPublicKeyInfo, &publicKeyInfo)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to decode SubjectPublicKeyInfo")
		}

		keyHash[h] = certutil.Digest(h, publicKeyInfo.PublicKey.RightAlign())
		nameHash[h] = certutil.Digest(h, bundle.Cert.RawSubject)

		logger.Infof("src=NewIssuer, label=%s, alg=%s, keyHash=%s, nameHash=%s",
			cfg.Label, certutil.HashAlgoToStr(h), hex.EncodeToString(keyHash[h]), hex.EncodeToString(nameHash[h]))
	}

	return &Issuer{
		skid:           certutil.GetSubjectKeyID(bundle.Cert),
		rkid:           certutil.GetSubjectKeyID(bundle.RootCert),
		signer:         cryptoSigner,
		bundle:         bundle,
		label:          cfg.Label,
		crlURL:         crl,
		aiaURL:         aia,
		ocspURL:        ocsp,
		crlNextUpdate:  cfg.CRLExpiry.TimeDuration(),
		ocspNextUpdate: cfg.OCSPExpiry.TimeDuration(),
		caCerts:        strings.TrimSpace(bundle.CertPEM) + "\n" + strings.TrimSpace(bundle.CACertsPEM),
		keyHash:        keyHash,
		nameHash:       nameHash,
	}, nil
}
