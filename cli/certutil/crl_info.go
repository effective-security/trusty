package certutil

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
)

// CRLInfoFlags specifies flags for CRLInfo action
type CRLInfoFlags struct {
	In *string
}

// CRLInfo shows CRL details
func CRLInfo(c ctl.Control, p interface{}) error {
	flags := p.(*CRLInfoFlags)

	// Load CRL
	der, err := c.(*cli.Cli).ReadFileOrStdin(*flags.In)
	if err != nil {
		return errors.Annotate(err, "unable to load CRL file")
	}

	crl, err := x509.ParseCRL(der)
	if err != nil {
		return errors.Annotate(err, "unable to prase CRL")
	}

	print.CertificateList(c.Writer(), crl)

	return nil
}

// CRLFetchFlags specifies flags for CRLFetch action
type CRLFetchFlags struct {
	CertFile *string
	OutDir   *string
	All      *bool
	Print    *bool
}

// CRLFetch shows CRL details
func CRLFetch(c ctl.Control, p interface{}) error {
	flags := p.(*CRLFetchFlags)
	isVerbose := c.(*cli.Cli).Verbose()
	w := c.Writer()

	// Load PEM
	pem, err := c.(*cli.Cli).ReadFileOrStdin(*flags.CertFile)
	if err != nil {
		return errors.Annotate(err, "unable to load PEM file")
	}

	list, err := certutil.ParseChainFromPEM(pem)
	if err != nil {
		return errors.Annotate(err, "unable to parse PEM")
	}

	if !*flags.All {
		// take only leaf cert
		list = list[:1]
	}

	for _, crt := range list {
		if len(crt.CRLDistributionPoints) < 1 {
			if isVerbose {
				fmt.Fprintf(w, "CRL DP is not present; CN=%q\n", crt.Subject.String())
			}
			continue
		}

		crldp := crt.CRLDistributionPoints[0]
		if isVerbose {
			fmt.Fprintf(w, "fetching CRL from %q\n", crldp)
		}

		body, err := download(crldp)
		if err != nil {
			return errors.Trace(err)
		}

		crl, err := x509.ParseCRL(body)
		if err != nil {
			return errors.Annotate(err, "unable to prase CRL")
		}
		if *flags.Print {
			print.CertificateList(c.Writer(), crl)
		}

		if *flags.OutDir != "" {
			filename := path.Join(*flags.OutDir, fmt.Sprintf("%s.crl", certutil.GetIssuerID(crt)))
			err = ioutil.WriteFile(filename, body, 0644)
			if err != nil {
				return errors.Annotatef(err, "unable to write CRL: %s", filename)
			}
		}
	}
	return nil
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to fetch from %s", url)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to download from %s", url)
	}

	return body, nil
}
