package authority

import (
	"crypto"

	"github.com/ekspand/trusty/pkg/csr"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
)

// NewRoot creates a new root certificate from the certificate request.
func NewRoot(profile string, cfg *Config, provider cryptoprov.Provider, req *csr.CertificateRequest) (certPEM, csrPEM, key []byte, err error) {
	err = req.Validate()
	if err != nil {
		err = errors.Annotate(err, "invalid request")
		return
	}

	var (
		gkey  crypto.PrivateKey
		keyID string
		c     = csr.NewProvider(provider)
	)

	csrPEM, gkey, keyID, err = c.GenerateKeyAndRequest(req)
	if err != nil {
		err = errors.Annotate(err, "process request")
		return
	}

	signer := gkey.(crypto.Signer)
	uri, keyBytes, err := provider.ExportKey(keyID)
	if err != nil {
		err = errors.Annotate(err, "failed to export key")
		return
	}

	if keyBytes == nil {
		key = []byte(uri)
	} else {
		key = keyBytes
	}

	issuer := &Issuer{
		cfg:     *cfg,
		signer:  signer,
		sigAlgo: csr.DefaultSigAlgo(signer),
	}
	err = issuer.cfg.Validate()
	if err != nil {
		err = errors.Annotate(err, "invalid configuration")
		return
	}

	sreq := csr.SignRequest{
		SAN:     req.SAN,
		Request: string(csrPEM),
		Profile: profile,
		Subject: &csr.X509Subject{
			CN:    req.CN,
			Names: req.Names,
		},
	}

	_, certPEM, err = issuer.Sign(sreq)
	if err != nil {
		err = errors.Annotate(err, "sign request")
		return
	}
	return
}
