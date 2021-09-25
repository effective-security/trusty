package status

import (
	"context"
	"fmt"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/pkg/print"
)

// Version shows the service version
func Version(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.StatusClient().Version(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.ServerVersion(c.Writer(), res)
	}
	return nil
}

// Server shows trusty server status
func Server(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.StatusClient().Server(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.ServerStatusResponse(c.Writer(), res)
	}

	return nil
}

// Caller shows the Caller status
func Caller(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.StatusClient().Caller(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.CallerStatusResponse(c.Writer(), res)
	}

	return nil
}
