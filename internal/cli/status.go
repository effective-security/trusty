package cli

import (
	"context"

	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/pkg/errors"
)

// VersionCmd shows the service version
type VersionCmd struct{}

// Run the command
func (a *VersionCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return err
	}
	defer client.Close()

	res, err := client.StatusClient().Version(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		_ = cli.WriteJSON(res)
	} else {
		print.ServerVersion(cli.Writer(), res)
	}
	return nil
}

// ServerStatusCmd shows the service status
type ServerStatusCmd struct{}

// Run the command
func (a *ServerStatusCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return err
	}
	defer client.Close()

	res, err := client.StatusClient().Server(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		_ = cli.WriteJSON(res)
	} else {
		print.ServerStatusResponse(cli.Writer(), res)
	}

	return nil
}

// CallerCmd shows the caller status
type CallerCmd struct{}

// Run the command
func (a *CallerCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.WFEServerName)
	if err != nil {
		return err
	}
	defer client.Close()

	res, err := client.StatusClient().Caller(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		_ = cli.WriteJSON(res)
	} else {
		print.CallerStatusResponse(cli.Writer(), res)
	}

	return nil
}
