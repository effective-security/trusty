package martini

import (
	"context"

	"github.com/ekspand/trusty/cli"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// UserProfile prints the current user info
func UserProfile(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.RefreshToken(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.CertificatesTable(c.Writer(), res.List)
		}
	*/
	return nil
}
