package status

import (
	"context"
	"fmt"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/trusty/cli"
	"github.com/juju/errors"
)

// Version shows the service version
func Version(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	res, err := cli.Client().Status.Version(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		fmt.Fprintf(c.Writer(), "%s\n", res.GetVersion())
	}
	return nil
}

// Server shows trusty server status
func Server(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	res, err := cli.Client().Status.Server(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	// TODO: cli.IsJSON() else print.Status
	ctl.WriteJSON(c.Writer(), res)
	fmt.Fprint(c.Writer(), "\n")

	return nil
}

// Caller shows the Caller status
func Caller(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	res, err := cli.Client().Status.Caller(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	// TODO: cli.IsJSON() else print.Status
	ctl.WriteJSON(c.Writer(), res)
	fmt.Fprint(c.Writer(), "\n")

	return nil
}
