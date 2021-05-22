package certutil

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
)

// CertInfoFlags specifies flags for CertInfo action
type CertInfoFlags struct {
	In        *string
	Out       *string
	NotAfter  *string
	NoExpired *bool
}

// CertInfo shows certs
func CertInfo(c ctl.Control, p interface{}) error {
	flags := p.(*CertInfoFlags)

	// Load PEM
	pem, err := c.(*cli.Cli).ReadFileOrStdin(*flags.In)
	if err != nil {
		return errors.Annotate(err, "unable to load PEM file")
	}

	list, err := certutil.ParseChainFromPEM(pem)
	if err != nil {
		return errors.Annotate(err, "unable to parse PEM")
	}

	now := time.Now().UTC()
	if flags.NoExpired != nil && *flags.NoExpired == true {
		list = filterByNotAfter(list, now)
	}

	if flags.NotAfter != nil && *flags.NotAfter != "" {
		d, err := time.ParseDuration(*flags.NotAfter)
		if err != nil {
			return errors.Annotate(err, "unable to parse --not-after")
		}
		list = filterByAfter(list, now.Add(d))
	}

	print.Certificates(c.Writer(), list)

	if *flags.Out != "" {
		f, err := os.OpenFile(*flags.Out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			return errors.Annotate(err, "unable to create file")
		}
		defer f.Close()

		certutil.EncodeToPEM(f, true, list...)
	}

	return nil
}

func filterByNotAfter(list []*x509.Certificate, notAfter time.Time) []*x509.Certificate {
	filtered := make([]*x509.Certificate, 0, len(list))
	for _, c := range list {
		if c.NotAfter.After(notAfter) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func filterByAfter(list []*x509.Certificate, notAfter time.Time) []*x509.Certificate {
	filtered := make([]*x509.Certificate, 0, len(list))
	for _, c := range list {
		if !c.NotAfter.After(notAfter) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// ValidateFlags specifies flags for Validate action
type ValidateFlags struct {
	Cert *string
	CA   *string
	Root *string
	Out  *string
}

// Validate cert chain
func Validate(c ctl.Control, p interface{}) error {
	flags := p.(*ValidateFlags)

	var err error
	var certBytes, cas []byte

	// set roots to empty
	roots := []byte("# empty Root bundle\n")

	certBytes, err = ioutil.ReadFile(*flags.Cert)
	if err != nil {
		return errors.Annotate(err, "unable to load cert")
	}

	if flags.CA != nil && *flags.CA != "" {
		cas, err = ioutil.ReadFile(*flags.CA)
		if err != nil {
			return errors.Annotate(err, "unable to load CA bundle")
		}
	}
	if flags.Root != nil && *flags.Root != "" {
		roots, err = ioutil.ReadFile(*flags.Root)
		if err != nil {
			return errors.Annotate(err, "unable to load Root bundle")
		}
	}

	w := c.Writer()
	bundle, bundleStatus, err := certutil.VerifyBundleFromPEM(certBytes, cas, roots)
	if err != nil {
		if crt, err2 := certutil.ParseFromPEM(certBytes); err2 == nil {
			print.Certificate(w, crt)
		}
		return errors.Annotate(err, "unable to verify certificate")
	}

	if bundleStatus.IsUntrusted() {
		fmt.Fprintf(w, "ERROR: The cert is untrusted\n")
	}

	chain := bundle.Chain
	if bundle.RootCert != nil {
		chain = append(chain, bundle.RootCert)
	}

	print.Certificates(w, chain)

	if len(bundleStatus.ExpiringSKIs) > 0 {
		fmt.Fprintf(w, "WARNING: Expiring SKI:\n")
		for _, ski := range bundleStatus.ExpiringSKIs {
			fmt.Fprintf(w, "  -- %s\n", ski)
		}
	}
	if len(bundleStatus.Untrusted) > 0 {
		fmt.Fprintf(w, "WARNING: Untrusted SKI:\n")
		for _, ski := range bundleStatus.Untrusted {
			fmt.Fprintf(w, "  -- %s\n", ski)
		}
	}

	if flags.Out != nil && *flags.Out != "" {
		pem := bundle.CertPEM + "\n" + bundle.CACertsPEM
		err = ioutil.WriteFile(*flags.Out, []byte(pem), 0664)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
