package martini

import (
	"context"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
)

// SearchCorpsFlags defines flags for SearchCorps command
type SearchCorpsFlags struct {
	Name         *string
	Jurisdiction *string
}

// SearchCorps prints the open corporates search result
func SearchCorps(c ctl.Control, p interface{}) error {
	flags := p.(*SearchCorpsFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.SearchCorps(context.Background(), *flags.Name, *flags.Jurisdiction)
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
