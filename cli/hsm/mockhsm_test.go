package hsm_test

import (
	"crypto"
	"crypto/elliptic"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ekspand/trusty/cli/hsm"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockHsmSuite struct {
	testsuite.Suite
}

func Test_mockHsmSuite(t *testing.T) {
	s := new(mockHsmSuite)
	suite.Run(t, s)
}

func (s *mockHsmSuite) Test_Slots() {
	// without KeyManager interface
	c, _ := cryptoprov.New(&mockedProvider{}, nil)
	s.Cli.WithCryptoProvider(c)

	err := s.Run(hsm.Slots, nil)
	s.Require().Error(err)
	s.Equal("unsupported command for this crypto provider", err.Error())

	// with mock
	mocked := &mockedFull{
		tokens: []slot{
			{
				slotID:       uint(1),
				description:  "d123",
				label:        "label123",
				manufacturer: "man123",
				model:        "model123",
				serial:       "serial123",
			},
		},
	}
	c, _ = cryptoprov.New(mocked, nil)
	s.Cli.WithCryptoProvider(c)

	mocked.On("EnumTokens", mock.Anything, mock.Anything).Times(1).Return(nil)
	mocked.On("EnumTokens", mock.Anything, mock.Anything).Times(1).Return(errors.New("unexpected error"))

	err = s.Run(hsm.Slots, nil)
	s.Require().NoError(err)
	s.HasText("Slot: 1\n  Description:  d123\n  Token serial: serial123\n  Token label:  label123\n")

	err = s.Run(hsm.Slots, nil)
	s.Require().Error(err)
	s.Equal("unable to list slots: unexpected error", err.Error())

	// assert that the expectations were met
	mocked.AssertExpectations(s.T())
}

func (s *mockHsmSuite) Test_KeyInfo() {
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

	// without KeyManager interface
	c, _ := cryptoprov.New(&mockedProvider{}, nil)
	s.Cli.WithCryptoProvider(c)

	err := s.Run(hsm.KeyInfo, &flags)
	s.Require().Error(err)
	s.Equal("unsupported command for this crypto provider", err.Error())
}

func (s *mockHsmSuite) Test_GenKey() {
	flags := hsm.GenKeyFlags{}

	// without KeyManager interface
	c, _ := cryptoprov.New(&mockedProvider{}, nil)
	s.Cli.WithCryptoProvider(c)

	err := s.Run(hsm.GenKey, &flags)
	s.Require().Error(err)
	s.Equal("unsupported purpose: \"\"", err.Error())

	purpose := "sign"
	flags = hsm.GenKeyFlags{
		Purpose: &purpose,
	}
	err = s.Run(hsm.GenKey, &flags)
	s.Require().Error(err)
	s.Equal("invalid algorithm: ", err.Error())

	creationTime := time.Now()
	mocked := &mockedFull{
		keys: map[uint][]keyInfo{
			uint(1): {
				{
					id:               "123",
					label:            "label123",
					typ:              "RSA",
					class:            "class",
					currentVersionID: "v124",
					creationTime:     &creationTime,
				},
			},
		},
	}
	pk := struct{}{}
	mocked.On("GenerateRSAKey", mock.Anything, mock.Anything, mock.Anything).Return(pk, nil)
	mocked.On("IdentifyKey", mock.Anything).Return("keyID", "lalel", nil)
	mocked.On("ExportKey", mock.Anything).Return("uri", []byte{1, 2, 3}, nil)

	c, _ = cryptoprov.New(mocked, nil)
	s.Cli.WithCryptoProvider(c)

	algo := "RSA"
	size := 2048
	label := "test*"
	flags = hsm.GenKeyFlags{
		Purpose: &purpose,
		Algo:    &algo,
		Size:    &size,
		Label:   &label,
	}
	err = s.Run(hsm.GenKey, &flags)
	s.Require().NoError(err)
}

func (s *mockHsmSuite) Test_LsKeyFlags() {
	token := ""
	serial := ""
	prefix := ""

	flags := hsm.LsKeyFlags{
		Token:  &token,
		Serial: &serial,
		Prefix: &prefix,
	}

	// without KeyManager interface
	c, _ := cryptoprov.New(&mockedProvider{}, nil)
	s.Cli.WithCryptoProvider(c)

	err := s.Run(hsm.Keys, &flags)
	s.Require().Error(err)
	s.Equal("unsupported command for this crypto provider", err.Error())

	// with keys and creationTime
	creationTime := time.Now()
	mocked := &mockedFull{
		tokens: []slot{
			{
				slotID:       uint(1),
				description:  "d123",
				label:        "label123",
				manufacturer: "man123",
				model:        "model123",
				serial:       "serial123-30589673",
			},
		},
		keys: map[uint][]keyInfo{
			uint(1): {
				{
					id:               "123",
					label:            "label123",
					typ:              "RSA",
					class:            "class",
					currentVersionID: "v124",
					creationTime:     &creationTime,
				},
				{
					id:               "with_error",
					label:            "with_error",
					typ:              "ECDSA",
					class:            "class",
					currentVersionID: "v1235",
					creationTime:     &creationTime,
				},
			},
		},
	}
	c, _ = cryptoprov.New(mocked, nil)
	s.Cli.WithCryptoProvider(c)

	mocked.On("EnumTokens", mock.Anything, mock.Anything).Times(2).Return(nil)
	mocked.On("EnumKeys", mock.Anything, mock.Anything, mock.Anything).Times(1).Return(nil)
	mocked.On("EnumKeys", mock.Anything, "with_error", mock.Anything).Times(1).Return(errors.New("unexpected error"))

	err = s.Run(hsm.Keys, &flags)
	s.Require().NoError(err)
	s.HasText("Slot: 1\n  Description:  d123\n  Token serial: serial123-30589673\n  Token label:  label123\n")
	s.HasText("CreationTime:")

	prefix = "with_error"
	flags.Prefix = &prefix
	err = s.Run(hsm.Keys, &flags)
	s.Require().Error(err)
	s.Equal("failed to list keys on slot 1: unexpected error", err.Error())

	// assert that the expectations were met
	mocked.AssertExpectations(s.T())
}

func (s *mockHsmSuite) Test_RmKey() {
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

	// without KeyManager interface
	c, _ := cryptoprov.New(&mockedProvider{}, nil)
	s.Cli.WithCryptoProvider(c)

	err := s.Run(hsm.RmKey, &flags)
	s.Require().Error(err)
	s.Equal("unsupported command for this crypto provider", err.Error())

	// with keys and creationTime
	creationTime := time.Now()
	mocked := &mockedFull{
		tokens: []slot{
			{
				slotID:       uint(1),
				description:  "d123",
				label:        "label123",
				manufacturer: "man124",
				model:        "model124",
				serial:       "serial123-729458762934",
			},
		},
		keys: map[uint][]keyInfo{
			uint(1): {
				{
					id:               "123",
					label:            "label123",
					typ:              "ECDSA",
					class:            "class",
					currentVersionID: "v1278",
					creationTime:     &creationTime,
				},
				{
					id:               "with_error",
					label:            "with_error",
					typ:              "RSA",
					class:            "class",
					currentVersionID: "v1239",
					creationTime:     &creationTime,
				},
			},
		},
	}
	c, _ = cryptoprov.New(mocked, nil)
	s.Cli.WithCryptoProvider(c)

	mocked.On("EnumTokens", mock.Anything, mock.Anything).Times(6).Return(nil)
	mocked.On("DestroyKeyPairOnSlot", mock.Anything, "with_error").Return(errors.New("access denied"))
	mocked.On("DestroyKeyPairOnSlot", mock.Anything, mock.Anything).Return(nil)

	mocked.On("EnumKeys", mock.Anything, "with_error", mock.Anything).Times(1).Return(errors.New("unexpected error"))
	mocked.On("EnumKeys", mock.Anything, mock.Anything, mock.Anything).Times(4).Return(nil)

	// by ID
	id = "with_error"
	prefix = ""
	flags.ID = &id
	flags.Prefix = &prefix
	err = s.Run(hsm.RmKey, &flags)
	s.Require().Error(err)
	s.Equal(`unable to destroy key "with_error" on slot 1: access denied`, err.Error())

	// by Prefix, with error on EnumKeys
	id = ""
	prefix = "with_error"
	flags.ID = &id
	flags.Prefix = &prefix
	err = s.Run(hsm.RmKey, &flags)
	s.Require().Error(err)
	s.Equal("failed to list keys on slot 1: unexpected error", err.Error())

	// by Prefix, no keys found
	id = ""
	prefix = "not_found"
	flags.ID = &id
	flags.Prefix = &prefix
	err = s.Run(hsm.RmKey, &flags)
	s.Require().NoError(err)
	s.HasNoText(`"no keys found with prefix: not_found\n"`)

	// by Prefix, no Confirmation
	id = ""
	prefix = "label123"
	flags.ID = &id
	flags.Prefix = &prefix
	err = s.Run(hsm.RmKey, &flags)
	s.HasText(`found 1 key(s) with prefix: label123`)
	s.Require().Error(err)
	s.Equal("unable to get a confirmation to destroy keys: ReadString failed: [EOF]", err.Error())
	s.HasNoText(`"no keys found with prefix: not_found\n"`)

	// by Prefix, force
	id = ""
	prefix = "label123"
	force = true
	flags.ID = &id
	flags.Prefix = &prefix
	flags.Force = &force
	err = s.Run(hsm.RmKey, &flags)
	s.HasText(`found 1 key(s) with prefix: label123`)
	s.Require().NoError(err)
	s.HasText(`destroyed key: 123`)

	// by Prefix, with Confirmation
	s.Cli.WithReader(strings.NewReader("y\n"))

	id = ""
	prefix = "label123"
	force = false
	flags.ID = &id
	flags.Prefix = &prefix
	flags.Force = &force
	err = s.Run(hsm.RmKey, &flags)
	s.HasText(`found 1 key(s) with prefix: label123`)
	s.Require().NoError(err)
	s.HasText(`destroyed key: 123`)

	// assert that the expectations were met
	mocked.AssertExpectations(s.T())
}

//
// Mock
//
type mockedProvider struct {
	mock.Mock
}

func (m *mockedProvider) GenerateRSAKey(label string, bits int, purpose int) (crypto.PrivateKey, error) {
	args := m.Called(label, bits, purpose)
	return args.Get(0).(crypto.PrivateKey), args.Error(1)
}

func (m *mockedProvider) GenerateECDSAKey(label string, curve elliptic.Curve) (crypto.PrivateKey, error) {
	args := m.Called(label, curve)
	return args.Get(0).(crypto.PrivateKey), args.Error(1)
}

func (m *mockedProvider) IdentifyKey(k crypto.PrivateKey) (keyID, label string, err error) {
	args := m.Called(k)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockedProvider) ExportKey(keyID string) (string, []byte, error) {
	args := m.Called(keyID)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

func (m *mockedProvider) GetKey(keyID string) (crypto.PrivateKey, error) {
	args := m.Called(keyID)
	return args.Get(0).(crypto.PrivateKey), args.Error(1)
}

func (m *mockedProvider) Manufacturer() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockedProvider) Model() string {
	args := m.Called()
	return args.String(0)
}

type slot struct {
	slotID       uint
	description  string
	label        string
	manufacturer string
	model        string
	serial       string
}

type keyInfo struct {
	id               string
	label            string
	typ              string
	class            string
	currentVersionID string
	creationTime     *time.Time
}
type mockedFull struct {
	mockedProvider

	tokens []slot
	keys   map[uint][]keyInfo
}

func (m *mockedFull) CurrentSlotID() uint {
	args := m.Called()
	return uint(args.Int(0))
}

func (m *mockedFull) EnumTokens(currentSlotOnly bool, slotInfoFunc func(slotID uint, description, label, manufacturer, model, serial string) error) error {
	args := m.Called(currentSlotOnly, slotInfoFunc)
	err := args.Error(0)
	if err == nil {
		for _, token := range m.tokens {
			err = slotInfoFunc(token.slotID, token.description, token.label, token.manufacturer, token.model, token.serial)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (m *mockedFull) EnumKeys(slotID uint, prefix string, keyInfoFunc func(id, label, typ, class, currentVersionID string, creationTime *time.Time) error) error {
	args := m.Called(slotID, prefix, keyInfoFunc)
	err := args.Error(0)
	if err == nil {
		for _, key := range m.keys[slotID] {
			if prefix == "" || strings.HasPrefix(key.label, prefix) {
				err = keyInfoFunc(key.id, key.label, key.typ, key.class, key.currentVersionID, key.creationTime)
				if err != nil {
					return err
				}
			}
		}
	}
	return err
}

func (m *mockedFull) DestroyKeyPairOnSlot(slotID uint, keyID string) error {
	args := m.Called(slotID, keyID)
	return args.Error(0)
}

func (m *mockedFull) FindKeyPairOnSlot(slotID uint, keyID, label string) (crypto.PrivateKey, error) {
	args := m.Called(slotID, keyID, label)
	return args.Get(0).(crypto.PrivateKey), args.Error(1)
}

func (m *mockedFull) KeyInfo(slotID uint, keyID string, includePublic bool, keyInfoFunc func(id, label, typ, class, currentVersionID, pubKey string, creationTime *time.Time) error) error {
	args := m.Called(slotID, keyID, includePublic, keyInfoFunc)
	return args.Error(0)
}

// ensure compiles
var _ cryptoprov.KeyManager = &mockedFull{}
