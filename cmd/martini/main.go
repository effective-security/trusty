package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/acme"
	"github.com/ekspand/trusty/cli/auth"
	"github.com/ekspand/trusty/cli/martini"
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

	// org register|validate|approve|subscribe|keys|members
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
	orgValidateFlags.OrgID = cmdValidateOrg.Flag("org", "organization ID").Required().String()

	orgSubscribeFlags := new(martini.CreateSubscriptionFlags)
	cmdSubscribeOrg := cmdOrgs.Command("subscribe", "create subscription").
		Action(cli.RegisterAction(martini.CreateSubscription, orgSubscribeFlags))
	orgSubscribeFlags.OrgID = cmdSubscribeOrg.Flag("org", "organization ID").Required().String()
	orgSubscribeFlags.CCName = cmdSubscribeOrg.Flag("cardholder", "CC cardholder").Required().String()
	orgSubscribeFlags.CCNumber = cmdSubscribeOrg.Flag("cc", "CC number").Required().String()
	orgSubscribeFlags.CCExpiry = cmdSubscribeOrg.Flag("expiry", "CC expiration date").Required().String()
	orgSubscribeFlags.CCCvv = cmdSubscribeOrg.Flag("cvv", "CC cvv").Required().String()
	orgSubscribeFlags.Years = cmdSubscribeOrg.Flag("years", "number of years").Required().Int()

	orgAPIKeysFlags := new(martini.APIKeysFlags)
	cmdOrgAPIKeys := cmdOrgs.Command("keys", "list API keys").
		Action(cli.RegisterAction(martini.APIKeys, orgAPIKeysFlags))
	orgAPIKeysFlags.OrgID = cmdOrgAPIKeys.Flag("org", "organization ID").Required().String()

	orgMembersFlags := new(martini.OrgMembersFlags)
	cmdOrgMembers := cmdOrgs.Command("members", "list members").
		Action(cli.RegisterAction(martini.OrgMembers, orgMembersFlags))
	orgMembersFlags.OrgID = cmdOrgMembers.Flag("org", "organization ID").Required().String()

	// acme account|order
	cmdAcme := app.Command("acme", "ACME operations").
		PreAction(cli.PopulateControl)

	acmeAccountFlags := new(acme.GetAccountFlags)
	cmdAcmeAccount := cmdAcme.Command("account", "show registered account").
		Action(cli.RegisterAction(acme.GetAccount, acmeAccountFlags))
	acmeAccountFlags.OrgID = cmdAcmeAccount.Flag("org", "organization ID").Required().String()
	acmeAccountFlags.EabMAC = cmdAcmeAccount.Flag("key", "EAB MAC key").Required().String()

	acmeRegisterAccountFlags := new(acme.RegisterAccountFlags)
	cmdAcmeRegister := cmdAcme.Command("register", "register account").
		Action(cli.RegisterAction(acme.RegisterAccount, acmeRegisterAccountFlags))
	acmeRegisterAccountFlags.OrgID = cmdAcmeRegister.Flag("org", "organization ID").Required().String()
	acmeRegisterAccountFlags.EabMAC = cmdAcmeRegister.Flag("key", "EAB MAC key").Required().String()
	acmeRegisterAccountFlags.Contact = cmdAcmeRegister.Flag("contact", "contact in mailto://name@org.com form").Required().Strings()

	acmeOrderFlags := new(acme.OrderFlags)
	cmdAcmeOrder := cmdAcme.Command("order", "order certificate").
		Action(cli.RegisterAction(acme.Order, acmeOrderFlags))
	acmeOrderFlags.OrgID = cmdAcmeOrder.Flag("org", "Organization ID").Required().String()
	acmeOrderFlags.SPC = cmdAcmeOrder.Flag("spc", "SPC file").Required().String()

	cli.Parse(args)
	return cli.ReturnCode()
}
