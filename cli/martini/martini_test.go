package martini_test

import (
	"testing"

	"github.com/ekspand/trusty/cli/martini"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	testsuite.Suite
}

func Test_CtlSuite(t *testing.T) {
	s := new(testSuite)
	s.WithFileServer()
	suite.Run(t, s)
}

func (s *testSuite) TestUserProfile() {
	err := s.Run(martini.UserProfile, nil)
	s.NoError(err)
	s.HasText("denis@ekspand.com")
}

func (s *testSuite) TestSearchCorps() {
	name := "peculiar ventures"
	jur := ""
	flags := martini.SearchCorpsFlags{
		Name:         &name,
		Jurisdiction: &jur,
	}

	err := s.Run(martini.SearchCorps, &flags)
	s.NoError(err)
	s.HasText("Private Limited Company")
}

func (s *testSuite) TestOrgs() {
	err := s.Run(martini.Orgs, nil)
	s.NoError(err)
	s.HasText(`"orgs": [`)
}

func (s *testSuite) TestCerts() {
	err := s.Run(martini.Certificates, nil)
	s.NoError(err)
	s.HasText(`"certificates": [`)
}

func (s *testSuite) TestFccFRN() {
	filer := "831188"
	flags := martini.FccFRNFlags{
		FilerID: &filer,
	}
	err := s.Run(martini.FccFRN, &flags)
	s.NoError(err)
	s.HasText(`"legal_name": "LOW LATENCY COMMUNICATIONS LLC"`)
}

func (s *testSuite) TestFccContact() {
	frn := "0024926677"
	flags := martini.FccContactFlags{
		FRN: &frn,
	}
	err := s.Run(martini.FccContact, &flags)
	s.NoError(err)
	s.HasText(`"contact_email": "mhardeman@lowlatencycomm.com"`)
}

func (s *testSuite) TestRegisterOrg() {
	filer := "123456"
	flags := martini.RegisterOrgFlags{
		FilerID: &filer,
	}
	err := s.Run(martini.RegisterOrg, &flags)
	s.NoError(err)
	s.HasText(`"status": "payment_pending",`)
}

// TODO: fix after the payment added
func (s *testSuite) TestApproveOrg() {
	code := "123456"
	token := "UZTBCIDb6j_aBpZf"
	flags := martini.ApprovergFlags{
		Token: &token,
		Code:  &code,
	}
	err := s.Run(martini.ApproveOrg, &flags)
	s.NoError(err)
	s.HasText(`"status": "approved"`)
}

func (s *testSuite) TestValidateOrg() {
	org := "82923411415760996"
	flags := martini.ValidateFlags{
		OrgID: &org,
	}
	err := s.Run(martini.ValidateOrg, &flags)
	s.NoError(err)
	s.HasText(`"status": "validation_pending",`)
}

func (s *testSuite) TestSubscribeOrg() {
	org := "82923411415760996"
	productID := "1234"
	flags := martini.CreateSubscriptionFlags{
		OrgID:     &org,
		ProductID: &productID,
	}
	err := s.Run(martini.CreateSubscription, &flags)
	s.NoError(err)
	s.HasText(`"status": "validation_pending"`)
}

func (s *testSuite) TestListSubscriptions() {
	err := s.Run(martini.Subscriptions, nil)
	s.NoError(err)
	s.HasText(`validation_pending`)
}

func (s *testSuite) TestListProducts() {
	err := s.Run(martini.Products, nil)
	s.NoError(err)
	s.HasText(`1 year subscription`)
}

func (s *testSuite) TestAPIKeys() {
	org := "82936648303640676"
	flags := martini.APIKeysFlags{
		OrgID: &org,
	}
	err := s.Run(martini.APIKeys, &flags)
	s.NoError(err)
	s.HasText(`"key": "_0zxP8c4AUrj_vnPmGXU_eEbA3AzkTXZ",`)
}

func (s *testSuite) TestOrgMembers() {
	org := "85334042257391942"
	flags := martini.OrgMembersFlags{
		OrgID: &org,
	}
	err := s.Run(martini.OrgMembers, &flags)
	s.NoError(err)
	s.HasText(`"name": "Denis Issoupov"`)
}

func (s *testSuite) TestDeleteOrg() {
	org := "123456"
	flags := martini.DeleteOrgFlags{
		OrgID: &org,
	}
	err := s.Run(martini.DeleteOrg, &flags)
	s.NoError(err)
}

func (s *testSuite) TestPayOrg() {
	stripeKey := "1234"
	clientSecret := "6789"
	noBrowser := true
	flags := martini.PayOrgFlags{
		StripeKey:    &stripeKey,
		ClientSecret: &clientSecret,
		NoBrowser:    &noBrowser,
	}
	err := s.Run(martini.PayOrg, &flags)
	s.NoError(err)
}
