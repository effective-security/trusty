package cli

import (
	"context"

	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

// CisCmd is the parent for cis command
type CisCmd struct {
	Roots GetRootsCmd `cmd:"" help:"list Root certificates"`
}

// GetRootsCmd defines flags for Roots command
type GetRootsCmd struct {
	Pem bool
}

// Run the command
func (a *GetRootsCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CISServerName)
	if err != nil {
		return err
	}
	defer client.Close()

	res, err := client.CIClient().GetRoots(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		print.Roots(cli.Writer(), res.Roots, a.Pem)
	}
	return nil
}
