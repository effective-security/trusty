package cli

import (
	"context"

	"github.com/effective-security/trusty/pkg/print"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
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

	res, err := client.GetRoots(context.Background(), &emptypb.Empty{})
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
