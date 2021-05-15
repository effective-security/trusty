package cis

import (
	"context"
	"fmt"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// GetRootsFlags defines flags for Roots command
type GetRootsFlags struct {
	Pem *bool
}

// Roots shows the root CAs
func Roots(c ctl.Control, p interface{}) error {
	flags := p.(*GetRootsFlags)

	cli := c.(*cli.Cli)
	client, err := cli.Client("cis")
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.CertInfoService.Roots(context.Background(), &pb.EmptyRequest{})
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
