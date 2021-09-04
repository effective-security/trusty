package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/acme"
	"github.com/ekspand/trusty/cli/status"
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

	app := ctl.NewApplication("macme", "Martini ACME client").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("macme %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
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

	// acme directory|account|register|order
	cmdAcme := app.Command("acme", "ACME operations").
		PreAction(cli.PopulateControl)

	cmdAcme.Command("directory", "show ACME directory").
		Action(cli.RegisterAction(acme.Directory, nil))

	acmeAccountFlags := new(acme.GetAccountFlags)
	cmdAcmeAccount := cmdAcme.Command("account", "show registered account").
		Action(cli.RegisterAction(acme.GetAccount, acmeAccountFlags))
	acmeAccountFlags.KeyID = cmdAcmeAccount.Flag("id", "key ID").Required().String()
	acmeAccountFlags.EabMAC = cmdAcmeAccount.Flag("key", "EAB MAC key").Required().String()

	acmeRegisterAccountFlags := new(acme.RegisterAccountFlags)
	cmdAcmeRegister := cmdAcme.Command("register", "register account").
		Action(cli.RegisterAction(acme.RegisterAccount, acmeRegisterAccountFlags))
	acmeRegisterAccountFlags.KeyID = cmdAcmeRegister.Flag("id", "key ID").Required().String()
	acmeRegisterAccountFlags.EabMAC = cmdAcmeRegister.Flag("key", "EAB MAC key").Required().String()
	acmeRegisterAccountFlags.Contact = cmdAcmeRegister.Flag("contact", "contact in mailto://name@org.com form").Required().Strings()

	acmeOrderFlags := new(acme.OrderFlags)
	cmdAcmeOrder := cmdAcme.Command("order", "order certificate").
		Action(cli.RegisterAction(acme.Order, acmeOrderFlags))
	acmeOrderFlags.KeyID = cmdAcmeOrder.Flag("id", "key ID").Required().String()
	acmeOrderFlags.SPC = cmdAcmeOrder.Flag("spc", "SPC file").Required().String()
	acmeOrderFlags.Days = cmdAcmeOrder.Flag("days", "validity period in days").Default("90").Int()

	cli.Parse(args)
	return cli.ReturnCode()
}
