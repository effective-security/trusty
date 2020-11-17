package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/ca"
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

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(out, "Unable to determine hostname: %v\n", err)
		hostname = "localhost"
	}

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithServiceCfg(),
		cli.WithHsmCfg(),
		cli.WithTLS(),
		cli.WithServer(fmt.Sprintf("%s:7891", hostname)),
	)
	defer cli.Close()

	app.Command("status", "show the server status").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(status.Server, nil))

	app.Command("version", "show the server version").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(status.Version, nil))

	app.Command("caller", "show the caller info").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(status.Caller, nil))

	cmdCA := app.Command("ca", "CA operations").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureClient)

	cmdCA.Command("issuers", "show the issuing CAs").
		Action(cli.RegisterAction(ca.Issuers, nil))

	cli.Parse(args)
	return cli.ReturnCode()
}
