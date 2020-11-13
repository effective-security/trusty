package csr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/config"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// GenCertFlags specifies flags for GenCert command
type GenCertFlags struct {
	SelfSign *bool
	// CACert specifies file name with CA cert
	CACert *string
	// CAKey specifies file name with CA key
	CAKey *string
	// CAConfig specifies file name with ca-config
	CAConfig *string
	// CsrProfile specifies file name with CSR profile
	CsrProfile *string
	// Label specifies name for generated key
	KeyLabel *string
	// SAN specifies coma separated alt names for generated cert
	SAN *string
	// Profile specifies the profile name from ca-config
	Profile *string
	// Output specifies the optional prefix for output files,
	// if not set, the output will be printed to STDOUT only
	Output *string
}

// GenCert generates a cert
func GenCert(c ctl.Control, p interface{}) error {
	flags := p.(*GenCertFlags)

	cryptoprov, defaultCrypto := c.(*cli.Cli).CryptoProv()
	if cryptoprov == nil {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	isscfg := &config.Issuer{}

	if flags.SelfSign != nil && *flags.SelfSign {
		if *flags.CACert != "" || *flags.CAKey != "" {
			return errors.Errorf("--self-sign can not be used with --ca-key")
		}
	} else {
		if *flags.CACert == "" || *flags.CAKey == "" {
			return errors.Errorf("CA certificate and key are required")
		}
		isscfg.CertFile = *flags.CACert
		isscfg.KeyFile = *flags.CAKey
	}

	// Load CSR
	csrf, err := c.(*cli.Cli).ReadFileOrStdin(*flags.CsrProfile)
	if err != nil {
		return errors.Annotate(err, "read CSR profile")
	}

	prov := csr.NewProvider(defaultCrypto)
	req := csr.CertificateRequest{
		KeyRequest: prov.NewKeyRequest(prefixKeyLabel(*flags.KeyLabel), "ECDSA", 256, csr.SigningKey),
	}

	err = json.Unmarshal(csrf, &req)
	if err != nil {
		return errors.Annotate(err, "invalid CSR profile")
	}

	// Load ca-config
	cacfg, err := authority.LoadConfig(*flags.CAConfig)
	if err != nil {
		return errors.Annotate(err, "ca-config")
	}
	err = cacfg.Validate()
	if err != nil {
		return errors.Annotate(err, "invalid ca-config")
	}

	var key, csrPEM, certPEM []byte

	if flags.SelfSign != nil && *flags.SelfSign {
		certPEM, csrPEM, key, err = authority.NewRoot(*flags.Profile,
			cacfg,
			defaultCrypto, &req)
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		issuer, err := authority.NewIssuer(isscfg, cacfg, cryptoprov)
		if err != nil {
			return errors.Annotate(err, "create issuer")
		}

		csrPEM, key, _, _, err = prov.CreateRequestAndExportKey(&req)
		if err != nil {
			key = nil
			return errors.Annotate(err, "process CSR")
		}

		var san []string
		if flags.SAN != nil && len(*flags.SAN) > 0 {
			san = strings.Split(*flags.SAN, ",")
		}
		signReq := csr.SignRequest{
			SAN:     san,
			Request: string(csrPEM),
			Profile: *flags.Profile,
		}

		_, certPEM, err = issuer.Sign(signReq)
		if err != nil {
			return errors.Annotate(err, "sign request")
		}
	}

	if *flags.Output == "" {
		PrintCert(c.Writer(), key, csrPEM, certPEM)
	} else {
		err = SaveCert(*flags.Output, key, csrPEM, certPEM)
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
