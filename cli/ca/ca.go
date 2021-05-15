package ca

import (
	"context"
	"fmt"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// Issuers shows the Issuing CAs
func Issuers(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.AuthorityService.Issuers(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Issuers(c.Writer(), res.Issuers, true)
	}
	return nil
}

// GetProfileFlags defines flags for Profile command
type GetProfileFlags struct {
	Profile *string
	Label   *string
}

// Profile shows the certifiate profile
func Profile(c ctl.Control, p interface{}) error {
	flags := p.(*GetProfileFlags)
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.AuthorityService.ProfileInfo(context.Background(), &pb.CertProfileInfoRequest{
		Profile: *flags.Profile,
		Label:   *flags.Label,
	})
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	fmt.Fprint(c.Writer(), "\n")
	// TODO: printer
	return nil
}
