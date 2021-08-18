package martini

import (
	"context"
	"fmt"

	v1 "github.com/ekspand/trusty/api/v1"
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

// Certificates prints the user's Certificates
func Certificates(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.Certificates(context.Background())
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

// ApprovergFlags defines flags for ApproveOrg command
type ApprovergFlags struct {
	Token  *string
	Code   *string
	Action *string
}

// ApproveOrg approves organization
func ApproveOrg(c ctl.Control, p interface{}) error {
	flags := p.(*ApprovergFlags)
	cl := c.(*cli.Cli)

	client, err := cl.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	action := cli.String(flags.Action)
	code := cli.String(flags.Code)
	if action == "approve" && code == "" {
		return errors.Errorf("approval action requires a code")
	}

	res, err := client.ApproveOrg(context.Background(),
		cli.String(flags.Token),
		code,
		action)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.ApproveOrgResponse(c.Writer(), res.List)
		}
	*/
	return nil
}

// ValidateFlags defines flags for ValidateOrg command
type ValidateFlags struct {
	OrgID *string
}

// ValidateOrg validates organization
func ValidateOrg(c ctl.Control, p interface{}) error {
	flags := p.(*ValidateFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.ValidateOrg(context.Background(), *flags.OrgID)
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

// APIKeysFlags defines flags for APIKeys command
type APIKeysFlags struct {
	OrgID *string
}

// APIKeys validates organization
func APIKeys(c ctl.Control, p interface{}) error {
	flags := p.(*APIKeysFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.GetOrgAPIKeys(context.Background(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.APIKeysResponse(c.Writer(), res.List)
		}
	*/
	return nil
}

// DeleteOrgFlags defines flags for DeleteOrg command
type DeleteOrgFlags struct {
	OrgID *string
}

// DeleteOrg deletes organization
func DeleteOrg(c ctl.Control, p interface{}) error {
	flags := p.(*DeleteOrgFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	err = client.DeleteOrg(context.Background(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Fprint(c.Writer(), "Deleted\n")

	return nil
}

// GetOrgFlags defines flags for GetOrg command
type GetOrgFlags struct {
	OrgID *string
}

// GetOrg returns organization
func GetOrg(c ctl.Control, p interface{}) error {
	flags := p.(*GetOrgFlags)
	cl := c.(*cli.Cli)

	client, err := cl.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	res, err := client.Org(context.Background(), cli.String(flags.OrgID))
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	/*
		if cli.IsJSON() {
			ctl.WriteJSON(c.Writer(), res)
			fmt.Fprint(c.Writer(), "\n")
		} else {
			print.Org(c.Writer(), res.List)
		}
	*/

	return nil
}
