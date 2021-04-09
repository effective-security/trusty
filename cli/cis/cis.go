package cis

import (
	"context"
	"fmt"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// GetRootsFlags defines flags for Roots command
type GetRootsFlags struct {
	OrgID *int64
	Pem   *bool
}

// Roots shows the root CAs
func Roots(c ctl.Control, p interface{}) error {
	flags := p.(*GetRootsFlags)

	cli := c.(*cli.Cli)
	res, err := cli.Client().CertInfoService.Roots(context.Background(), &trustypb.GetRootsRequest{
		OrgID: *flags.OrgID,
	})
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
