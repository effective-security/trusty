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
			print.UserProfile(c.Writer(), res.List)
		}
	*/
	return nil
}

// Orgs prints the user's Orgs
func Orgs(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.Orgs(context.Background())
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

// RegisterOrgFlags defines flags for RegisterOrg command
type RegisterOrgFlags struct {
	FilerID *string
}

// RegisterOrg starts registration flow
func RegisterOrg(c ctl.Control, p interface{}) error {
	flags := p.(*RegisterOrgFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.RegisterOrg(context.Background(), *flags.FilerID)
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

// ValidateOrgFlags defines flags for RegisterOrg command
type ValidateOrgFlags struct {
	Token *string
	Code  *string
}

// ValidateOrg validates organization
func ValidateOrg(c ctl.Control, p interface{}) error {
	flags := p.(*ValidateOrgFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.ValidateOrg(context.Background(), *flags.Token, *flags.Code)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.ValidateOrgResponse(c.Writer(), res.List)
		}
	*/
	return nil
}
