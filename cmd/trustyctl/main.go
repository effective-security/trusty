package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/effective-security/trusty/internal/cli"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/xpki/x/ctl"
)

type app struct {
	cli.Cli

	Version cli.VersionCmd      `cmd:"" help:"print remote server version"`
	Status  cli.ServerStatusCmd `cmd:"" help:"print remote server status"`
	Caller  cli.CallerCmd       `cmd:"" help:"print identity of the current user"`

	Ca  cli.CaCmd  `cmd:"" help:"CA commands"`
	Cis cli.CisCmd `cmd:"" help:"CIS commands"`
}

func main() {
	realMain(os.Args, os.Stdout, os.Stderr, os.Exit)
}

func realMain(args []string, out io.Writer, errout io.Writer, exit func(int)) {
	cl := app{
		Cli: cli.Cli{
			Version: ctl.VersionFlag("0.0.1"),
		},
	}
	cl.Cli.WithErrWriter(errout).
		WithWriter(out)

	parser, err := kong.New(&cl,
		kong.Name("trustyctl"),
		kong.Description("CTL for trusty server"),
		//kong.UsageOnError(),
		kong.Writers(out, errout),
		kong.Exit(exit),
		ctl.BoolPtrMapper,
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": version.Current().String(),
		})
	if err != nil {
		panic(err)
	}

	cli, err := parser.Parse(args[1:])
	parser.FatalIfErrorf(err)

	if cli != nil {
		err = cli.Run(&cl.Cli)
		cli.FatalIfErrorf(err)
	}
}
