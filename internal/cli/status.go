package cli

import (
	"google.golang.org/protobuf/types/known/emptypb"
)

// VersionCmd shows the service version
type VersionCmd struct{}

// Run the command
func (a *VersionCmd) Run(cli *Cli) error {
	client, err := cli.StatusClient()
	if err != nil {
		return err
	}

	res, err := client.Version(cli.Context(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	_ = cli.Print(res)
	return nil
}

// ServerStatusCmd shows the service status
type ServerStatusCmd struct{}

// Run the command
func (a *ServerStatusCmd) Run(cli *Cli) error {
	client, err := cli.StatusClient()
	if err != nil {
		return err
	}

	res, err := client.Server(cli.Context(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	_ = cli.Print(res)

	return nil
}

// CallerCmd shows the caller status
type CallerCmd struct{}

// Run the command
func (a *CallerCmd) Run(cli *Cli) error {
	client, err := cli.StatusClient()
	if err != nil {
		return err
	}

	res, err := client.Caller(cli.Context(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	_ = cli.Print(res)

	return nil
}
