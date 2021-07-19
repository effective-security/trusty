package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/martini"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
)

func main() {
	// Logs are set to os.Stderr, while output to os.Stdout
	rc := realMain(os.Args, os.Stdout, os.Stderr)
	os.Exit(int(rc))
}

func realMain(args []string, out io.Writer, errout io.Writer) ctl.ReturnCode {
	formatter := xlog.NewColorFormatter(errout, true)
	xlog.SetFormatter(formatter)

	app := ctl.NewApplication("martinictl", "A command-line utility for Martini API.").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("martinictl %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithServiceCfg(),
		cli.WithTLS(),
		cli.WithServer(""),
	)
	defer cli.Close()

	// opencorps
	searchCorpsFlags := new(martini.SearchCorpsFlags)
	cmdSearchCorps := app.Command("opencorps", "Search open corporations").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.SearchCorps, searchCorpsFlags))
	searchCorpsFlags.Name = cmdSearchCorps.Flag("name", "corporation name to search").Required().String()
	searchCorpsFlags.Jurisdiction = cmdSearchCorps.Flag("jur", "jurisdition code: us, us_wa, etc").String()

	cli.Parse(args)
	return cli.ReturnCode()
}
