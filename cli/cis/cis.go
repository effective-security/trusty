package cis

import (
	"context"
	"fmt"

	"github.com/go-phorce/dolly/ctl"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/pkg/print"
)

// GetRootsFlags defines flags for Roots command
type GetRootsFlags struct {
	Pem *bool
}

// Roots shows the root CAs
func Roots(c ctl.Control, p interface{}) error {
	flags := p.(*GetRootsFlags)

	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CISServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.CIClient().GetRoots(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Roots(c.Writer(), res.Roots, *flags.Pem)
	}
	return nil
}
