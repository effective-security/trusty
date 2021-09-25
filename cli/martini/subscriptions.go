package martini

import (
	"context"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/cli"
)

// CreateSubscriptionFlags defines flags for CreateSubscription command
type CreateSubscriptionFlags struct {
	OrgID     *string
	ProductID *string
}

// CreateSubscription pays for organization
func CreateSubscription(c ctl.Control, p interface{}) error {
	flags := p.(*CreateSubscriptionFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	req := &v1.CreateSubscriptionRequest{
		OrgID:     *flags.OrgID,
		ProductID: *flags.ProductID,
	}

	res, err := client.CreateSubscription(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)

	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.OrgResponse(c.Writer(), res.List)
		}
	*/
	return nil
}

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
