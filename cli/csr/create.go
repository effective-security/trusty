package csr

import (
	"encoding/json"
	"io/ioutil"

	"github.com/go-phorce/dolly/ctl"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/martinisecurity/trusty/pkg/print"
	"github.com/pkg/errors"
)

// CreateFlags specifies flags for Create command
type CreateFlags struct {
	// CsrProfile specifies file name with CSR profile
	CsrProfile *string
	// KeyLabel specifies name for generated key
	KeyLabel *string
	// Output specifies the optional prefix for output files,
	// if not set, the output will be printed to STDOUT only
	Output *string
}

// Create generates a key and creates a CSR
func Create(c ctl.Control, p interface{}) error {
	flags := p.(*CreateFlags)

	cryptoprov, defaultCrypto := c.(*cli.Cli).CryptoProv()
	if cryptoprov == nil {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	prov := csr.NewProvider(defaultCrypto)

	csrf, err := c.(*cli.Cli).ReadFileOrStdin(*flags.CsrProfile)
	if err != nil {
		return errors.WithMessage(err, "read CSR profile")
	}

	req := csr.CertificateRequest{
		KeyRequest: prov.NewKeyRequest(prefixKeyLabel(*flags.KeyLabel), "ECDSA", 256, csr.SigningKey),
	}

	err = json.Unmarshal(csrf, &req)
	if err != nil {
		return errors.WithMessage(err, "invalid CSR")
	}

	var key, csrPEM []byte
	csrPEM, key, _, _, err = prov.CreateRequestAndExportKey(&req)
	if err != nil {
		key = nil
		return errors.WithMessage(err, "process CSR")
	}

	if *flags.Output == "" {
		print.CSRandCert(c.Writer(), key, csrPEM, nil)
	} else {
		err = SaveCert(*flags.Output, key, csrPEM, nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// SaveCert to file
func SaveCert(baseName string, key, csrPEM, certPEM []byte) error {
	var err error
	if len(certPEM) > 0 {
		err = ioutil.WriteFile(baseName+".pem", certPEM, 0664)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(csrPEM) > 0 {
		err = ioutil.WriteFile(baseName+".csr", csrPEM, 0664)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(key) > 0 {
		err = ioutil.WriteFile(baseName+".key", key, 0600)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
