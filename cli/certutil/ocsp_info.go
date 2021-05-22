package certutil

import (
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"golang.org/x/crypto/ocsp"
)

// OCSPInfoFlags specifies flags for CRLInfo action
type OCSPInfoFlags struct {
	In     *string
	Issuer *string
}

// OCSPInfo shows OCSP response details
func OCSPInfo(c ctl.Control, p interface{}) error {
	flags := p.(*OCSPInfoFlags)

	// Load DER
	der, err := c.(*cli.Cli).ReadFileOrStdin(*flags.In)
	if err != nil {
		return errors.Annotate(err, "unable to load OCSP file")
	}

	res, err := ocsp.ParseResponse(der, nil)
	if err != nil {
		return errors.Annotate(err, "unable to prase OCSP")
	}

	print.OCSPResponse(c.Writer(), res)

	return nil
}
