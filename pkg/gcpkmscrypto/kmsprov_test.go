package gcpkmscrypto_test

import (
	"context"
	"crypto"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/googleapis/gax-go/v2"
	"github.com/martinisecurity/trusty/pkg/gcpkmscrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Test_KmsProvider(t *testing.T) {
	cfg := &mockTokenCfg{
		manufacturer: "GCPKMS-roots",
		model:        "GCPKMS",
		atts:         "Keyring=projects/trusty-dev-319216/locations/us-central1/keyRings/trusty-dev",
	}

	original := gcpkmscrypto.KmsClientFactory
	defer func() {
		gcpkmscrypto.KmsClientFactory = original
	}()

	mocked := &mockedProvider{
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
		keys: []*kmspb.CryptoKey{
			{
				Name:       "123",
				CreateTime: timestamppb.Now(),
			},
		},
	}
	gcpkmscrypto.KmsClientFactory = func() (gcpkmscrypto.KmsClient, error) {
		return mocked, nil
	}

	it := &kms.CryptoKeyIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) (results []*kmspb.CryptoKey, nextPageToken string, err error) {
		return mocked.keys, "", nil
	}

	key := &kmspb.CryptoKey{
		Name:            "key1",
		Labels:          map[string]string{"label": "testKey"},
		VersionTemplate: &kmspb.CryptoKeyVersionTemplate{},
	}

	pubKey := &kmspb.PublicKey{
		Name: "key1",
		Pem: `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpSwQTzpTI9LFgLtdHAMHl0oEIgwf
2i7YbXYfrucbC0xPekQgsxEJJqQJauSwOugli7FYYKxyapk3j6lGImVCbA==
-----END PUBLIC KEY-----`,
	}

	crc := wrapperspb.Int64Value{
		Value: int64(gcpkmscrypto.Crc32c([]byte(`bytes`))),
	}
	sig := &kmspb.AsymmetricSignResponse{
		Signature:            []byte(`bytes`),
		VerifiedDigestCrc32C: true,
		SignatureCrc32C:      &crc,
	}

	mocked.On("EnumTokens", mock.Anything, mock.Anything).Return(nil)
	mocked.On("EnumKeys", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mocked.On("ListCryptoKeys", mock.Anything, mock.Anything, mock.Anything).Return(it)
	mocked.On("CreateCryptoKey", mock.Anything, mock.Anything, mock.Anything).Return(key, nil)
	mocked.On("GetCryptoKey", mock.Anything, mock.Anything, mock.Anything).Return(key, nil)
	mocked.On("GetPublicKey", mock.Anything, mock.Anything, mock.Anything).Return(pubKey, nil)
	mocked.On("AsymmetricSign", mock.Anything, mock.Anything, mock.Anything).Return(sig, nil)
	mocked.On("Close").Return(nil)

	prov, err := gcpkmscrypto.KmsLoader(cfg)
	require.NoError(t, err)
	require.NotNil(t, prov)

	assert.Equal(t, "GCPKMS-roots", prov.Manufacturer())
	assert.Equal(t, "GCPKMS", prov.Model())

	//	count := 0
	mgr := prov.(cryptoprov.KeyManager)

	mgr.EnumTokens(false, func(slotID uint, description, label, manufacturer, model, serial string) error {
		assert.Equal(t, "GCPKMS-roots", manufacturer)
		assert.Equal(t, "GCPKMS", model)
		return nil
	})

	// TODO: GCP KMS makes it very hard to mock keys iterator
	/*
		err = mgr.EnumKeys(mgr.CurrentSlotID(), "", func(id, label, typ, class, currentVersionID string, creationTime *time.Time) error {
			count++
			return nil
		})
		require.NoError(t, err)
	*/
	rsacases := []struct {
		size int
		hash crypto.Hash
	}{
		{2048, crypto.SHA256},
		{3072, crypto.SHA256},
		{4096, crypto.SHA512},
	}

	for _, tc := range rsacases {
		pvk, err := prov.GenerateRSAKey(fmt.Sprintf("RSA_%d_%s", tc.size, guid.MustCreate()), tc.size, 1)
		require.NoError(t, err)

		keyID, _, err := prov.IdentifyKey(pvk)
		require.NoError(t, err)

		uri, _, err := prov.ExportKey(keyID)
		require.NoError(t, err)
		assert.Contains(t, uri, "pkcs11:manufacturer=GCPKMS-roots;")

		signer := pvk.(crypto.Signer)
		require.NotNil(t, signer)

		hash := tc.hash.New()
		digest := hash.Sum([]byte(`digest`))
		_, err = signer.Sign(rand.Reader, digest[:hash.Size()], tc.hash)
		require.NoError(t, err)
	}

	eccases := []struct {
		curve elliptic.Curve
		hash  crypto.Hash
	}{
		{elliptic.P256(), crypto.SHA256},
		{elliptic.P384(), crypto.SHA384},
		//{elliptic.P521(), crypto.SHA512},
	}

	for _, tc := range eccases {
		pvk, err := prov.GenerateECDSAKey(fmt.Sprintf("ECC_%s", guid.MustCreate()), tc.curve)
		require.NoError(t, err)

		keyID, _, err := prov.IdentifyKey(pvk)
		require.NoError(t, err)

		_, err = prov.GetKey(keyID)
		require.NoError(t, err)

		signer := pvk.(crypto.Signer)
		require.NotNil(t, signer)

		hash := tc.hash.New()
		digest := hash.Sum([]byte(`digest`))
		_, err = signer.Sign(rand.Reader, digest[:hash.Size()], tc.hash)
		require.NoError(t, err)

		err = mgr.KeyInfo(mgr.CurrentSlotID(), keyID, true, func(id, label, typ, class, currentVersionID, publey string, creationTime *time.Time) error {
			return nil
		})
		require.NoError(t, err)
	}

	/*
		addedCount := 0
		err = mgr.EnumKeys(mgr.CurrentSlotID(), "", func(id, label, typ, class, currentVersionID string, creationTime *time.Time) error {
			addedCount++

			mgr.DestroyKeyPairOnSlot(mgr.CurrentSlotID(), id)
			return nil
		})
		require.NoError(t, err)
	*/

	_, err = mgr.FindKeyPairOnSlot(0, "123412", "")
	require.Error(t, err)

	closer := prov.(io.Closer)
	err = closer.Close()
	require.NoError(t, err)
}

func TestKeyLabelOrID(t *testing.T) {
	l1, k1 := gcpkmscrypto.KeyLabelAndID("plain")
	l2, k2 := gcpkmscrypto.KeyLabelAndID("plain")
	assert.Equal(t, l1, l2)
	assert.NotEqual(t, k1, k2)
	assert.Equal(t, "plain", l1)
	assert.Equal(t, "plain", l2)

	l1, k1 = gcpkmscrypto.KeyLabelAndID("plain*")
	l2, k2 = gcpkmscrypto.KeyLabelAndID("plain*")
	assert.Equal(t, l1, l2)
	assert.NotEqual(t, k1, k2)
	assert.Equal(t, "plain", l1)
	assert.Equal(t, "plain", l2)
}

//
// mockTokenCfg
//

type mockTokenCfg struct {
	manufacturer string
	model        string
	path         string
	tokenSerial  string
	tokenLabel   string
	pin          string
	atts         string
}

// Manufacturer name of the manufacturer
func (m *mockTokenCfg) Manufacturer() string {
	return m.manufacturer
}

// Model name of the device
func (m *mockTokenCfg) Model() string {
	return m.model
}

// Full path to PKCS#11 library
func (m *mockTokenCfg) Path() string {
	return m.path
}

// Token serial number
func (m *mockTokenCfg) TokenSerial() string {
	return m.tokenSerial
}

// Token label
func (m *mockTokenCfg) TokenLabel() string {
	return m.tokenLabel
}

// Pin is a secret to access the token.
// If it's prefixed with `file:`, then it will be loaded from the file.
func (m *mockTokenCfg) Pin() string {
	return m.pin
}

// Comma separated key=value pair of attributes(e.g. "ServiceName=x,UserName=y")
func (m *mockTokenCfg) Attributes() string {
	return m.atts
}

type slot struct {
	slotID       uint
	description  string
	label        string
	manufacturer string
	model        string
	serial       string
}

type mockedProvider struct {
	mock.Mock

	tokens []slot
	keys   []*kmspb.CryptoKey
}

func (m *mockedProvider) EnumTokens(currentSlotOnly bool, slotInfoFunc func(slotID uint, description, label, manufacturer, model, serial string) error) error {
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

func (m *mockedProvider) ListCryptoKeys(ctx context.Context, req *kmspb.ListCryptoKeysRequest, ops ...gax.CallOption) *kms.CryptoKeyIterator {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kms.CryptoKeyIterator)
}

func (m *mockedProvider) GetCryptoKey(ctx context.Context, req *kmspb.GetCryptoKeyRequest, ops ...gax.CallOption) (*kmspb.CryptoKey, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.CryptoKey), args.Error(1)
}

func (m *mockedProvider) GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, ops ...gax.CallOption) (*kmspb.PublicKey, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.PublicKey), args.Error(1)
}

func (m *mockedProvider) GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, ops ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.CryptoKeyVersion), args.Error(1)
}

func (m *mockedProvider) DestroyCryptoKeyVersion(ctx context.Context, req *kmspb.DestroyCryptoKeyVersionRequest, ops ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.CryptoKeyVersion), args.Error(1)
}

func (m *mockedProvider) AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, ops ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.AsymmetricSignResponse), args.Error(1)
}

func (m *mockedProvider) CreateCryptoKey(ctx context.Context, req *kmspb.CreateCryptoKeyRequest, ops ...gax.CallOption) (*kmspb.CryptoKey, error) {
	args := m.Called(ctx, req, ops)
	return args.Get(0).(*kmspb.CryptoKey), args.Error(1)
}

func (m *mockedProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}
