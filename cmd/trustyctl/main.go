package main

import (
	"fmt"
	"io"
	"os"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/cli/ca"
	"github.com/martinisecurity/trusty/cli/cis"
	"github.com/martinisecurity/trusty/cli/status"
	"github.com/martinisecurity/trusty/internal/version"
)

func main() {
	// Logs are set to os.Stderr, while output to os.Stdout
	rc := realMain(os.Args, os.Stdout, os.Stderr)
	os.Exit(int(rc))
}

func realMain(args []string, out io.Writer, errout io.Writer) ctl.ReturnCode {
	formatter := xlog.NewColorFormatter(errout, true)
	xlog.SetFormatter(formatter)

	app := ctl.NewApplication("trustyctl", "A command-line utility for controlling Trusty.").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("trustyctl %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithServiceCfg(),
		cli.WithHsmCfg(),
		cli.WithTLS(),
		cli.WithServer(""),
	)
	defer cli.Close()

	app.Command("status", "show the server status").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Server, nil))

	app.Command("version", "show the server version").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Version, nil))

	app.Command("caller", "show the caller info").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Caller, nil))

	// ca: issuers|profile|sign|publish_crl|certificate

	cmdCA := app.Command("ca", "CA operations").
		PreAction(cli.PopulateControl)

	listIssuersFlags := new(ca.ListIssuersFlags)
	cmdListIssuers := cmdCA.Command("issuers", "show the issuing CAs").
		Action(cli.RegisterAction(ca.ListIssuers, listIssuersFlags))
	listIssuersFlags.Limit = cmdListIssuers.Flag("limit", "limit the items in response").Int64()
	listIssuersFlags.After = cmdListIssuers.Flag("after", "pagination after ID").Uint64()
	listIssuersFlags.Bundle = cmdListIssuers.Flag("bundle", "return PEM of the chain").Bool()

	getProfileFlags := new(ca.GetProfileFlags)
	profileCmd := cmdCA.Command("profile", "show the certificate profile").
		Action(cli.RegisterAction(ca.Profile, getProfileFlags))
	getProfileFlags.Label = profileCmd.Flag("label", "profile label").String()

	signFlags := new(ca.SignFlags)
	signCmd := cmdCA.Command("sign", "sign CSR").
		Action(cli.RegisterAction(ca.Sign, signFlags))
	signFlags.Request = signCmd.Flag("csr", "CSR file").Required().String()
	signFlags.Profile = signCmd.Flag("profile", "certificate profile").Required().String()
	signFlags.IssuerLabel = signCmd.Flag("issuer", "label of issuer to use").String()
	signFlags.Token = signCmd.Flag("token", "authorization token for the request").String()
	signFlags.Label = signCmd.Flag("label", "certificate label").String()
	signFlags.SAN = signCmd.Flag("san", "optional SAN").Strings()
	signFlags.Out = signCmd.Flag("out", "output file name").String()

	listCertsFlags := new(ca.ListCertsFlags)
	listCertsCmd := cmdCA.Command("certs", "print the certificates").
		Action(cli.RegisterAction(ca.ListCerts, listCertsFlags))
	listCertsFlags.Ikid = listCertsCmd.Flag("ikid", "Issuer Key Identifier").Required().String()
	listCertsFlags.Limit = listCertsCmd.Flag("limit", "max limit of the certificates to print").Int()
	listCertsFlags.After = listCertsCmd.Flag("after", "the certificate ID for pagination").String()

	certLabelFlags := new(ca.UpdateCertLabelFlags)
	certLabelCmd := cmdCA.Command("label", "update the certificate label").
		Action(cli.RegisterAction(ca.UpdateCertLabel, certLabelFlags))
	certLabelFlags.ID = certLabelCmd.Flag("id", "certificate ID").Uint64()
	certLabelFlags.Label = certLabelCmd.Flag("label", "certificates label").String()

	rlistCertsFlags := new(ca.ListCertsFlags)
	revokedCmd := cmdCA.Command("revoked", "print the revoked certificates").
		Action(cli.RegisterAction(ca.ListRevokedCerts, rlistCertsFlags))
	rlistCertsFlags.Ikid = revokedCmd.Flag("ikid", "Issuer Key Identifier").Required().String()
	rlistCertsFlags.Limit = revokedCmd.Flag("limit", "max limit of the certificates to print").Int()
	rlistCertsFlags.After = revokedCmd.Flag("after", "the certificate ID for pagination").String()

	publishCrlFlags := new(ca.PublishCrlsFlags)
	publishCrlCmd := cmdCA.Command("publish_crl", "publish CRL").
		Action(cli.RegisterAction(ca.PublishCrls, publishCrlFlags))
	publishCrlFlags.Ikid = publishCrlCmd.Flag("ikid", "Issuer Key Identifier").Required().String()

	revokeFlags := new(ca.RevokeFlags)
	revokeCmd := cmdCA.Command("revoke", "revoke a certificate").
		Action(cli.RegisterAction(ca.Revoke, revokeFlags))
	revokeFlags.ID = revokeCmd.Flag("id", "ID of the certificate").Uint64()
	revokeFlags.SKID = revokeCmd.Flag("skid", "Subject Key Identifier").String()
	revokeFlags.IKID = revokeCmd.Flag("ikid", "Issuer Key Identifier").String()
	revokeFlags.Serial = revokeCmd.Flag("sn", "Serial Number").String()
	revokeFlags.Reason = revokeCmd.Flag("reason", "Reason for revocation").Int()

	getCertFlags := new(ca.GetCertificateFlags)
	getCertCmd := cmdCA.Command("certificate", "get a certificate").
		Action(cli.RegisterAction(ca.GetCertificate, getCertFlags))
	getCertFlags.ID = getCertCmd.Flag("id", "ID of the certificate").Uint64()
	getCertFlags.SKID = getCertCmd.Flag("skid", "Subject Key Identifier").String()

	// cis: roots|certs|revoked

	cmdCIS := app.Command("cis", "CIS operations").
		PreAction(cli.PopulateControl)

	getRootsFlags := new(cis.GetRootsFlags)
	rootsCmd := cmdCIS.Command("roots", "show the roots").
		Action(cli.RegisterAction(cis.Roots, getRootsFlags))
	getRootsFlags.Pem = rootsCmd.Flag("pem", "specifies to print PEM").Bool()

	cli.Parse(args)
	return cli.ReturnCode()
}
