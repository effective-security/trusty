package cli

import (
	"context"

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
	client, err := cli.CISClient()
	if err != nil {
		return err
	}

	res, err := client.GetRoots(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.WithStack(err)
	}

	if a.Pem {
		print.Roots(cli.Writer(), res.Roots, true)
	} else {
		_ = cli.Print(res)
	}

	return nil
}
