package cis

import (
	"context"
	"fmt"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/golang/protobuf/ptypes/empty"
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

// ListCertsFlags defines flags for ListCerts command
type ListCertsFlags struct {
	Ikid  *string
	Limit *int
	After *string
}

// ListCerts prints the certifiates
func ListCerts(c ctl.Control, p interface{}) error {
	flags := p.(*ListCertsFlags)
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CISServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	after := uint64(0)
	if *flags.After != "" {
		after, err = db.ID(*flags.After)
		if err != nil {
			return errors.Annotate(err, "unable to parse --after")
		}
	}

	res, err := client.CIClient().ListCertificates(context.Background(), &pb.ListByIssuerRequest{
		Ikid:  *flags.Ikid,
		Limit: int64(*flags.Limit),
		After: after,
	})
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.CertificatesTable(c.Writer(), res.List)
	}

	return nil
}

// ListRevokedCerts prints the revoked certifiates
func ListRevokedCerts(c ctl.Control, p interface{}) error {
	flags := p.(*ListCertsFlags)
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CISServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	after := uint64(0)
	if *flags.After != "" {
		after, err = db.ID(*flags.After)
		if err != nil {
			return errors.Annotate(err, "unable to parse --after")
		}
	}

	res, err := client.CIClient().ListRevokedCertificates(context.Background(), &pb.ListByIssuerRequest{
		Ikid:  *flags.Ikid,
		Limit: int64(*flags.Limit),
		After: after,
	})
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.RevokedCertificatesTable(c.Writer(), res.List)
	}

	return nil
}
