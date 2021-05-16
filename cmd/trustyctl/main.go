package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/auth"
	"github.com/ekspand/trusty/cli/ca"
	"github.com/ekspand/trusty/cli/cis"
	"github.com/ekspand/trusty/cli/status"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/cmd", "trustyctl")

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

	// login
	loginFlags := new(auth.AuthenticateFlags)
	cmdLogin := app.Command("login", "login to Trusty").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(auth.Authenticate, loginFlags))
	loginFlags.NoBrowser = cmdLogin.Flag("no-browser", "disable openning in browser").Bool()

	// ca: issuers|profile|sign

	cmdCA := app.Command("ca", "CA operations").
		PreAction(cli.PopulateControl)

	cmdCA.Command("issuers", "show the issuing CAs").
		Action(cli.RegisterAction(ca.Issuers, nil))

	getProfileFlags := new(ca.GetProfileFlags)
	profileCmd := cmdCA.Command("profile", "show the certificate profile").
		Action(cli.RegisterAction(ca.Profile, getProfileFlags))
	getProfileFlags.Profile = profileCmd.Flag("name", "profile name").Required().String()
	getProfileFlags.Label = profileCmd.Flag("issuer", "issuer label").String()

	signFlags := new(ca.SignFlags)
	signCmd := cmdCA.Command("sign", "sign CSR").
		Action(cli.RegisterAction(ca.Sign, signFlags))
	signFlags.Request = signCmd.Flag("csr", "CSR file").Required().String()
	signFlags.Profile = signCmd.Flag("profile", "certificate profile").Required().String()
	signFlags.IssuerLabel = signCmd.Flag("issuer", "label of issuer to use").String()
	signFlags.Token = signCmd.Flag("token", "authorization token for the request").String()
	signFlags.SAN = signCmd.Flag("san", "optional SAN").Strings()
	signFlags.Out = signCmd.Flag("out", "output file name").String()

	// cis: roots

	cmdCIS := app.Command("cis", "CIS operations").
		PreAction(cli.PopulateControl)

	getRootsFlags := new(cis.GetRootsFlags)
	rootsCmd := cmdCIS.Command("roots", "show the roots").
		Action(cli.RegisterAction(cis.Roots, getRootsFlags))
	getRootsFlags.Pem = rootsCmd.Flag("pem", "specifies to print PEM").Bool()

	cli.Parse(args)
	return cli.ReturnCode()
}
