package main

import (
	"fmt"
	"io"
	"os"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/trusty/cli"
	"github.com/go-phorce/trusty/cli/csr"
	"github.com/go-phorce/trusty/version"
)

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty/cmd", "trusty-tool")

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
		cli.WithHsmCfg(),
	)
	defer cli.Close()

	// csr create-root
	cmdCSR := app.Command("csr", "CSR commands").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureCryptoProvider)

	createRootFlags := new(csr.RootFlags)
	cmdCreateRoot := cmdCSR.Command("create-root", "Create self-signed CA certificate").
		Action(cli.RegisterAction(csr.Root, createRootFlags))

	createRootFlags.CsrProfile = cmdCreateRoot.Flag("csr-profile", "CSR profile file").Required().String()
	createRootFlags.KeyLabel = cmdCreateRoot.Flag("key-label", "Label for generated key").String()
	createRootFlags.Output = cmdCreateRoot.Flag("out", "specifies the optional prefix for output files").String()

	cli.Parse(args)
	return cli.ReturnCode()
}
