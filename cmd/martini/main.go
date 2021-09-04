package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/auth"
	"github.com/ekspand/trusty/cli/martini"
	"github.com/ekspand/trusty/cli/status"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
)

func main() {
	// Logs are set to os.Stderr, while output to os.Stdout
	rc := realMain(os.Args, os.Stdout, os.Stderr)
	os.Exit(int(rc))
}

func realMain(args []string, out io.Writer, errout io.Writer) ctl.ReturnCode {
	formatter := xlog.NewColorFormatter(errout, true)
	xlog.SetFormatter(formatter)

	app := ctl.NewApplication("martini", "A command-line utility for Martini API.").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("martini %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithTLS(),
		cli.WithServer(""),
	)
	defer cli.Close()

	app.Command("status", "show the server status").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Server, nil))

	app.Command("version", "show the server version").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Version, nil))

	app.Command("caller", "show the caller info").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(status.Caller, nil))

	// login
	prov := "google"
	loginFlags := &auth.AuthenticateFlags{
		Provider: &prov,
	}
	app.Command("login", "login and obtain authorization token").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(auth.Authenticate, loginFlags))

	// user info
	app.Command("userinfo", "show the user profile").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.UserProfile, nil))

	// user orgs
	app.Command("orgs", "show the user's orgs").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.Orgs, nil))

	// user certs
	app.Command("certificates", "show the user's certificates").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.Certificates, nil))

	// user subscriptions
	app.Command("subscriptions", "show the user's subscriptions").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.Subscriptions, nil))

	// products
	app.Command("products", "show the available products").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.Products, nil))

	// opencorps
	searchCorpsFlags := new(martini.SearchCorpsFlags)
	cmdSearchCorps := app.Command("opencorps", "search open corporations").
		PreAction(cli.PopulateControl).
		Action(cli.RegisterAction(martini.SearchCorps, searchCorpsFlags))
	searchCorpsFlags.Name = cmdSearchCorps.Flag("name", "corporation name to search").Required().String()
	searchCorpsFlags.Jurisdiction = cmdSearchCorps.Flag("jur", "jurisdition code: us, us_wa, etc").String()

	// fcc frn
	cmdFCC := app.Command("fcc", "FCC operations").
		PreAction(cli.PopulateControl)

	fccFRNFlags := new(martini.FccFRNFlags)
	cmdFRN := cmdFCC.Command("frn", "returns FRN for filer").
		Action(cli.RegisterAction(martini.FccFRN, fccFRNFlags))
	fccFRNFlags.FilerID = cmdFRN.Flag("filer", "filer ID").Required().String()

	fccContactFlags := new(martini.FccContactFlags)
	cmdFccContact := cmdFCC.Command("contact", "returns contact for FRN").
		Action(cli.RegisterAction(martini.FccContact, fccContactFlags))
	fccContactFlags.FRN = cmdFccContact.Flag("frn", "FRN to query").Required().String()

	// org register|validate|approve|subscribe|keys|members|delete|get|search
	cmdOrgs := app.Command("org", "Orgs operations").
		PreAction(cli.PopulateControl)

	orgRegisterFlags := new(martini.RegisterOrgFlags)
	cmdRegisterOrg := cmdOrgs.Command("register", "registers organization").
		Action(cli.RegisterAction(martini.RegisterOrg, orgRegisterFlags))
	orgRegisterFlags.FilerID = cmdRegisterOrg.Flag("filer", "filer ID").Required().String()

	orgApproveFlags := new(martini.ApprovergFlags)
	cmdApproveOrg := cmdOrgs.Command("approve", "approve organization validation").
		Action(cli.RegisterAction(martini.ApproveOrg, orgApproveFlags))
	orgApproveFlags.Token = cmdApproveOrg.Flag("token", "approver's token").Required().String()
	orgApproveFlags.Code = cmdApproveOrg.Flag("code", "requestor's code").Required().String()
	approve := "approve"
	orgApproveFlags.Action = &approve

	orgDenyFlags := new(martini.ApprovergFlags)
	cmdDenyOrg := cmdOrgs.Command("deny", "deny organization validation").
		Action(cli.RegisterAction(martini.ApproveOrg, orgDenyFlags))
	orgDenyFlags.Token = cmdDenyOrg.Flag("token", "approver's token").Required().String()
	deny := "deny"
	orgDenyFlags.Action = &deny

	orgInfoFlags := new(martini.ApprovergFlags)
	cmdInfoOrg := cmdOrgs.Command("info", "info organization request").
		Action(cli.RegisterAction(martini.ApproveOrg, orgInfoFlags))
	orgInfoFlags.Token = cmdInfoOrg.Flag("token", "approver's token").Required().String()
	info := "info"
	orgInfoFlags.Action = &info

	orgValidateFlags := new(martini.ValidateFlags)
	cmdValidateOrg := cmdOrgs.Command("validate", "approve organization validation").
		Action(cli.RegisterAction(martini.ValidateOrg, orgValidateFlags))
	orgValidateFlags.OrgID = cmdValidateOrg.Flag("id", "organization ID").Required().String()

	subscribeFlags := new(martini.CreateSubscriptionFlags)
	cmdSubscribe := cmdOrgs.Command("subscribe", "subscribe to org").
		Action(cli.RegisterAction(martini.CreateSubscription, subscribeFlags))
	subscribeFlags.OrgID = cmdSubscribe.Flag("id", "organization ID").Required().String()
	subscribeFlags.ProductID = cmdSubscribe.Flag("product", "product id to subscribe to").Required().String()

	orgAPIKeysFlags := new(martini.APIKeysFlags)
	cmdOrgAPIKeys := cmdOrgs.Command("keys", "list API keys").
		Action(cli.RegisterAction(martini.APIKeys, orgAPIKeysFlags))
	orgAPIKeysFlags.OrgID = cmdOrgAPIKeys.Flag("id", "organization ID").Required().String()

	orgMembersFlags := new(martini.OrgMembersFlags)
	cmdOrgMembers := cmdOrgs.Command("members", "list members").
		Action(cli.RegisterAction(martini.OrgMembers, orgMembersFlags))
	orgMembersFlags.OrgID = cmdOrgMembers.Flag("id", "organization ID").Required().String()

	orgDeleteFlags := new(martini.DeleteOrgFlags)
	cmdOrgDelete := cmdOrgs.Command("delete", "delete organization").
		Action(cli.RegisterAction(martini.DeleteOrg, orgDeleteFlags))
	orgDeleteFlags.OrgID = cmdOrgDelete.Flag("id", "organization ID").Required().String()

	orgGetFlags := new(martini.GetOrgFlags)
	cmdOrgGet := cmdOrgs.Command("get", "show the organization").
		Action(cli.RegisterAction(martini.GetOrg, orgGetFlags))
	orgGetFlags.OrgID = cmdOrgGet.Flag("id", "organization ID").Required().String()

	orgPayFlags := new(martini.PayOrgFlags)
	cmdOrgPay := cmdOrgs.Command("pay", "pay for org").
		Action(cli.RegisterAction(martini.PayOrg, orgPayFlags))
	orgPayFlags.StripeKey = cmdOrgPay.Flag("stripe-key", "Stripe publishable key").Required().String()
	orgPayFlags.ClientSecret = cmdOrgPay.Flag("client-secret", "client secret after subscription is created").Required().String()
	orgPayFlags.NoBrowser = cmdOrgPay.Flag("no-browser", "disable openning in browser").Bool()

	orgSearchFlags := new(martini.SearchOrgsFlags)
	cmdOrgSearch := cmdOrgs.Command("search", "search organization").
		Action(cli.RegisterAction(martini.SearchOrgs, orgSearchFlags))
	orgSearchFlags.FRN = cmdOrgSearch.Flag("frn", "FRN").String()
	orgSearchFlags.FillerID = cmdOrgSearch.Flag("filler", "FCC 499 ID").String()

	cli.Parse(args)
	return cli.ReturnCode()
}
