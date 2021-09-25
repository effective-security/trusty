package martini

import (
	"context"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
)

// OrgMembersFlags defines flags for OrgMembers command
type OrgMembersFlags struct {
	OrgID *string
}

// OrgMembers prints the org members
func OrgMembers(c ctl.Control, p interface{}) error {
	flags := p.(*OrgMembersFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.OrgMembers(context.Background(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.OrgMembers(c.Writer(), res.List)
		}
	*/
	return nil
}

// OrgMemberAddFlags defines flags for OrgMemberAdd command
type OrgMemberAddFlags struct {
	OrgID *string
	Email *string
	Role  *string
}

// OrgMemberAdd adds a member to an org
func OrgMemberAdd(c ctl.Control, p interface{}) error {
	flags := p.(*OrgMemberAddFlags)
	cl := c.(*cli.Cli)

	client, err := cl.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.OrgMemberAdd(context.Background(),
		cli.String(flags.OrgID),
		cli.String(flags.Email),
		cli.String(flags.Role))
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.OrgMembers(c.Writer(), res.List)
		}
	*/
	return nil
}

// OrgMemberRemoveFlags defines flags for OrgMemberRemove command
type OrgMemberRemoveFlags struct {
	OrgID  *string
	UserID *string
}

// OrgMemberRemove removes a member from an org
func OrgMemberRemove(c ctl.Control, p interface{}) error {
	flags := p.(*OrgMemberRemoveFlags)
	cl := c.(*cli.Cli)

	client, err := cl.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}
	client.WithAuthorization()

	res, err := client.OrgMemberRemove(context.Background(),
		cli.String(flags.OrgID),
		cli.String(flags.UserID))
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.OrgMembers(c.Writer(), res.List)
		}
	*/
	return nil
}
