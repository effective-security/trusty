package main

import (
	"fmt"
	"io"
	"os"

	"github.com/effective-security/xlog"
	"github.com/go-phorce/dolly/ctl"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/cli/certutil"
	"github.com/martinisecurity/trusty/internal/version"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/cmd", "trusty-tool")

const (
	rcError   = 1
	rcSuccess = 0
)

func main() {
	// Logs are set to os.Stderr, while output to os.Stdout
	rc := realMain(os.Args, os.Stdout, os.Stderr)
	os.Exit(int(rc))
}

func realMain(args []string, out io.Writer, errout io.Writer) ctl.ReturnCode {
	formatter := xlog.NewColorFormatter(errout, true)
	xlog.SetFormatter(formatter)

	app := ctl.NewApplication("trusty-tool", "A command-line utility for issuing offline certificates.").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("trusty-tool %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithHsmCfg(), cli.WithPlainKey(),
	)
	defer cli.Close()

	// csr self-sign|create|sign|gencert
	cmdCSR := app.Command("csr", "CSR commands").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureCryptoProvider)
	/*
		createRootFlags := new(csr.RootFlags)
		cmdCreateRoot := cmdCSR.Command("self-sign", "generates key and creates self-signed certificate").
			Action(cli.RegisterAction(csr.Root, createRootFlags))
		createRootFlags.CsrProfile = cmdCreateRoot.Flag("csr-profile", "CSR profile file").Required().String()
		createRootFlags.KeyLabel = cmdCreateRoot.Flag("key-label", "label for generated key").String()
		createRootFlags.Output = cmdCreateRoot.Flag("out", "specifies the optional prefix for output files").String()
	*/

	csrinfoFlags := new(certutil.CSRInfoFlags)
	cmdCSRInfo := cmdCSR.Command("info", "Print certificate request info").
		Action(cli.RegisterAction(certutil.CSRInfo, csrinfoFlags))
	csrinfoFlags.In = cmdCSRInfo.Flag("in", "PEM-encoded file with certificate request").Required().String()

	// cert info|validate
	cmdCert := app.Command("cert", "Cert utils").
		PreAction(cli.PopulateControl)

	certinfoFlags := new(certutil.CertInfoFlags)
	cmdCertInfo := cmdCert.Command("info", "Print certificate info").
		Action(cli.RegisterAction(certutil.CertInfo, certinfoFlags))
	certinfoFlags.In = cmdCertInfo.Flag("in", "PEM-encoded file with certificates").Required().String()
	certinfoFlags.Out = cmdCertInfo.Flag("out", "Optional, output PEM-encoded file").String()
	certinfoFlags.NotAfter = cmdCertInfo.Flag("not-after", "Optional filter by Not After in duration format: 1h").String()
	certinfoFlags.NoExpired = cmdCertInfo.Flag("no-expired", "Optional filter for expired certs").Bool()

	validateFlags := new(certutil.ValidateFlags)
	cmdCertValidate := cmdCert.Command("validate", "Validate certificate chain").
		Action(cli.RegisterAction(certutil.Validate, validateFlags))
	validateFlags.Cert = cmdCertValidate.Flag("cert", "PEM-encoded file with certificate").Required().String()
	validateFlags.CA = cmdCertValidate.Flag("ca", "PEM-encoded file with intermediate CA").String()
	validateFlags.Root = cmdCertValidate.Flag("root", "PEM-encoded file with Root CA").String()
	validateFlags.Out = cmdCertValidate.Flag("out", "Output PEM-encoded file with Leaf and intermediate CA certificates").String()

	// crl info|get
	cmdCRL := app.Command("crl", "CRL utils").
		PreAction(cli.PopulateControl)

	crlinfoFlags := new(certutil.CRLInfoFlags)
	cmdCRLInfo := cmdCRL.Command("info", "Print certificate request info").
		Action(cli.RegisterAction(certutil.CRLInfo, crlinfoFlags))
	crlinfoFlags.In = cmdCRLInfo.Flag("in", "DER-encoded CRL file").Required().String()

	crlFetchFlags := new(certutil.CRLFetchFlags)
	cmdCrlFetch := cmdCRL.Command("get", "Download CRL").
		Action(cli.RegisterAction(certutil.CRLFetch, crlFetchFlags))
	crlFetchFlags.CertFile = cmdCrlFetch.Flag("cert", "PEM-encoded certificate").Short('c').Required().String()
	crlFetchFlags.OutDir = cmdCrlFetch.Flag("out", "Folder to strore, optional").Short('o').String()
	crlFetchFlags.All = cmdCrlFetch.Flag("all", "Process all certificates in the bundle").Short('a').Bool()
	crlFetchFlags.Print = cmdCrlFetch.Flag("print", "Print CRL info").Short('p').Bool()

	// ocsp info|get
	cmdOCSP := app.Command("ocsp", "OCSP utils").
		PreAction(cli.PopulateControl)

	ocspinfoFlags := new(certutil.OCSPInfoFlags)
	cmdOCSPInfo := cmdOCSP.Command("info", "Print OCSP response info").
		Action(cli.RegisterAction(certutil.OCSPInfo, ocspinfoFlags))
	ocspinfoFlags.In = cmdOCSPInfo.Flag("in", "DER-encoded OCSP file").Required().String()

	cli.Parse(args)
	return cli.ReturnCode()
}
