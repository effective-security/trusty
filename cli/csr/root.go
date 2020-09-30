package csr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	cfsslcli "github.com/cloudflare/cfssl/cli"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/csrprov"
	"github.com/go-phorce/trusty/cli"
	"github.com/juju/errors"
)

// RootFlags specifies flags for Root command
type RootFlags struct {
	// CsrProfile specifies file name with CSR profile
	CsrProfile *string
	// KeyLabel specifies name for generated key
	KeyLabel *string
	// Output specifies the optional prefix for output files,
	// if not set, the output will be printed to STDOUT only
	Output *string
}

// Root generates a self-signed cert
func Root(c ctl.Control, p interface{}) error {
	flags := p.(*RootFlags)

	cryptoprov := c.(*cli.Cli).CryptoProv()
	if cryptoprov == nil {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	prov := csrprov.New(cryptoprov.Default())

	csrf, err := c.(*cli.Cli).ReadFileOrStdin(*flags.CsrProfile)
	if err != nil {
		return errors.Annotate(err, "read CSR profile")
	}

	req := csrprov.CertificateRequest{
		// TODO: alg and size from params
		KeyRequest: prov.NewKeyRequest(prefixKeyLabel(*flags.KeyLabel), "ECDSA", 256, csrprov.Signing),
	}

	err = json.Unmarshal(csrf, &req)
	if err != nil {
		return errors.Annotate(err, "invalid CSR")
	}

	var key, csrPEM, cert []byte
	cert, csrPEM, key, err = prov.NewRoot(&req)
	if err != nil {
		return errors.Annotate(err, "init CA")
	}

	if *flags.Output == "" {
		cfsslcli.PrintCert(key, csrPEM, cert)
	} else {
		baseName := *flags.Output

		err = ioutil.WriteFile(baseName+".pem", cert, 0664)
		if err != nil {
			return errors.Trace(err)
		}
		err = ioutil.WriteFile(baseName+".csr", csrPEM, 0664)
		if err != nil {
			return errors.Trace(err)
		}
		err = ioutil.WriteFile(baseName+"-key.pem", key, 0600)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// prefixKeyLabel adds a date prefix to label for a key
func prefixKeyLabel(label string) string {
	if strings.HasSuffix(label, "*") {
		g := guid.MustCreate()
		t := time.Now().UTC()
		label = strings.TrimSuffix(label, "*") +
			fmt.Sprintf("_%04d%02d%02d%02d%02d%02d_%x", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), g[:4])
	}

	return label
}
