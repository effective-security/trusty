package martini

import (
	"context"

	"github.com/ekspand/trusty/cli"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// FccFRNFlags defines flags for FccFRN command
type FccFRNFlags struct {
	FilerID *string
}

// FccFRN returns FRN
func FccFRN(c ctl.Control, p interface{}) error {
	flags := p.(*FccFRNFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.FccFRN(context.Background(), *flags.FilerID)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.FRN(c.Writer(), res.List)
		}
	*/
	return nil
}

// FccContactFlags defines flags for FccContact command
type FccContactFlags struct {
	FRN *string
}

// FccContact returns Contact
func FccContact(c ctl.Control, p interface{}) error {
	flags := p.(*FccContactFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.FccContact(context.Background(), *flags.FRN)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.FccContact(c.Writer(), res.List)
		}
	*/
	return nil
}
