package status

import (
	"context"
	"fmt"

	"github.com/go-phorce/dolly/ctl"
	pb "github.com/go-phorce/trusty/api/v1/serverpb"
	"github.com/go-phorce/trusty/cli"
	"github.com/juju/errors"
)

// Version shows the service version
func Version(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	client := pb.NewStatusClient(cli.GrpcConnection())
	res, err := client.Version(context.Background(), &pb.EmptyRequest{})
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		fmt.Fprintf(c.Writer(), "%s\n", res.GetVersion())
	}
	return nil
}

// Server shows trusty server status
func Server(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)
	client := pb.NewStatusClient(cli.GrpcConnection())
	res, err := client.Server(context.Background(), &pb.EmptyRequest{})
	if err != nil {
		return errors.Trace(err)
	}

	// TODO: cli.IsJSON() else print.Status
	ctl.WriteJSON(c.Writer(), res)
	fmt.Fprint(c.Writer(), "\n")

	return nil
}
