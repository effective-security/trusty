package martini

import (
	"context"

	"github.com/ekspand/trusty/cli"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// Subscriptions prints the user's Subscriptions
func Subscriptions(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.ListSubscriptions(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.Orgs(c.Writer(), res.List)
		}
	*/
	return nil
}

// Products prints the available products
func Products(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.ListSubscriptionsProducts(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.Orgs(c.Writer(), res.List)
		}
	*/
	return nil
}
