package hsm_test

import (
	"testing"

	"github.com/ekspand/trusty/cli/hsm"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/stretchr/testify/suite"
)

type softHsmSuite struct {
	testsuite.Suite
}

func Test_SoftHsmSuite(t *testing.T) {
	s := new(softHsmSuite)

	s.WithHSM()
	s.WithAppFlags([]string{"--hsm-cfg", "/tmp/trusty/softhsm/unittest_hsm.json"})
	suite.Run(t, s)
}

func (s *softHsmSuite) SetupSuite() {
	s.Suite.SetupSuite()
	err := s.Cli.EnsureCryptoProvider()
	s.Require().NoError(err)
}

func (s *softHsmSuite) Test_Slots() {
	err := s.Run(hsm.Slots, nil)
	s.NoError(err)
	s.HasText(`Slot:`)
}

func (s *softHsmSuite) Test_KeyInfo() {
	token := ""
	serial := ""
	id := ""
	withPub := true

	flags := hsm.KeyInfoFlags{
		Token:  &token,
		Serial: &serial,
		ID:     &id,
		Public: &withPub,
	}

	// 1. with default
	err := s.Run(hsm.KeyInfo, &flags)
	s.NoError(err)
	s.HasText(`CurrentVersionID:`)

	// 2. with non-existing Key ID
	id = "123456vlwbeoivbwerfvbwefvwfev"
	flags.ID = &id
	err = s.Run(hsm.KeyInfo, &flags)
	s.NoError(err)
	s.HasText(`failed to get key info on slot`)

	// 3. with non-existing Serial
	serial = "123456vlwbeoivbwe46747"
	id = ""
	flags.ID = &id
	flags.Serial = &serial
	err = s.Run(hsm.KeyInfo, &flags)
	s.NoError(err)
	s.HasText(`no slots found with serial`)
}

func (s *softHsmSuite) Test_LsKeyFlags() {
	token := ""
	serial := ""
	prefix := ""

	flags := hsm.LsKeyFlags{
		Token:  &token,
		Serial: &serial,
		Prefix: &prefix,
	}

	// 1. with default
	err := s.Run(hsm.Keys, &flags)
	s.NoError(err)
	s.HasText(`CurrentVersionID:`)

	// 2. with prefix
	prefix = "123456vlwbeo6959579wefvwfev"
	flags.Prefix = &prefix
	err = s.Run(hsm.Keys, &flags)
	s.NoError(err)
	s.HasText(`no keys found with prefix: 123456vlwbeo6959579wefvwfev`)
}

func (s *softHsmSuite) Test_RmKey() {
	token := ""
	serial := ""
	id := ""
	prefix := ""
	force := false

	flags := hsm.RmKeyFlags{
		Token:  &token,
		Serial: &serial,
		ID:     &id,
		Prefix: &prefix,
		Force:  &force,
	}

	// with default
	err := s.Run(hsm.RmKey, &flags)
	s.Require().Error(err)
	s.Equal("either of --prefix and --id must be specified", err.Error())

	// with mutual exclusive flags
	id = "123456vlwbeoivbwerfvbwefvwfev"
	prefix = "123456vlwbeoivbwerfvbwefvwfev"
	flags.ID = &id
	flags.Prefix = &prefix

	err = s.Run(hsm.RmKey, &flags)
	s.Require().Error(err)
	s.Equal("--prefix and --id should not be specified together", err.Error())

	// with ID
	id = "123456vlwbeoivbwerfvbwefvwfev"
	prefix = ""
	flags.ID = &id
	flags.Prefix = &prefix

	err = s.Run(hsm.RmKey, &flags)
	s.Require().NoError(err)
	s.HasText(`destroyed key: 123456vlwbeoivbwerfvbwefvwfev`)

	// with prefix
	id = ""
	prefix = "58576857856785678567" // non existing
	flags.ID = &id
	flags.Prefix = &prefix

	err = s.Run(hsm.RmKey, &flags)
	s.Require().NoError(err)
	s.HasText("no keys found with prefix: 58576857856785678567\n")
	s.HasNoText(` Type 'yes' to continue or 'no' to cancel. [y/n]:`)
	s.HasNoText(`destroyed key:`)
}
