package certutil

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/pkg/print"
)

// CSRInfoFlags specifies flags for CSRInfo action
type CSRInfoFlags struct {
	In *string
}

// CSRInfo shows certs
func CSRInfo(c ctl.Control, p interface{}) error {
	flags := p.(*CSRInfoFlags)

	// Load CSR
	csrb, err := c.(*cli.Cli).ReadFileOrStdin(*flags.In)
	if err != nil {
		return errors.Annotate(err, "unable to load CSR file")
	}

	block, _ := pem.Decode(csrb)
	if block == nil || "CERTIFICATE REQUEST" != block.Type {
		return errors.New("invalid CSR file")
	}

	csrv, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return errors.Annotate(err, "unable to prase CSR")
	}

	print.CertificateRequest(c.Writer(), csrv)

	return nil
}
