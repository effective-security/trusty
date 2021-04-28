package csr

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
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
		return errors.Annotate(err, "read CSR profile")
	}

	req := csr.CertificateRequest{
		KeyRequest: prov.NewKeyRequest(prefixKeyLabel(*flags.KeyLabel), "ECDSA", 256, csr.SigningKey),
	}

	err = json.Unmarshal(csrf, &req)
	if err != nil {
		return errors.Annotate(err, "invalid CSR")
	}

	var key, csrPEM []byte
	csrPEM, key, _, _, err = prov.CreateRequestAndExportKey(&req)
	if err != nil {
		key = nil
		return errors.Annotate(err, "process CSR")
	}

	if *flags.Output == "" {
		print.CSRandCert(c.Writer(), key, csrPEM, nil)
	} else {
		err = SaveCert(*flags.Output, key, csrPEM, nil)
		if err != nil {
			return errors.Trace(err)
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
			return errors.Trace(err)
		}
	}
	if len(csrPEM) > 0 {
		err = ioutil.WriteFile(baseName+".csr", csrPEM, 0664)
		if err != nil {
			return errors.Trace(err)
		}
	}
	if len(key) > 0 {
		err = ioutil.WriteFile(baseName+"-key.pem", key, 0600)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}
