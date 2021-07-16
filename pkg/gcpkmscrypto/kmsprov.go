package gcpkmscrypto

import (
	"context"
	"crypto"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path"
	"strings"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/googleapis/gax-go/v2"
	"github.com/juju/errors"
	"google.golang.org/api/iterator"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "gcpkmscrypto")

// ProviderName specifies a provider name
const ProviderName = "GCPKMS"

// KmsClient interface
type KmsClient interface {
	ListCryptoKeys(context.Context, *kmspb.ListCryptoKeysRequest, ...gax.CallOption) *kms.CryptoKeyIterator
	GetCryptoKey(context.Context, *kmspb.GetCryptoKeyRequest, ...gax.CallOption) (*kmspb.CryptoKey, error)
	GetPublicKey(context.Context, *kmspb.GetPublicKeyRequest, ...gax.CallOption) (*kmspb.PublicKey, error)
	GetCryptoKeyVersion(context.Context, *kmspb.GetCryptoKeyVersionRequest, ...gax.CallOption) (*kmspb.CryptoKeyVersion, error)
	DestroyCryptoKeyVersion(context.Context, *kmspb.DestroyCryptoKeyVersionRequest, ...gax.CallOption) (*kmspb.CryptoKeyVersion, error)
	AsymmetricSign(context.Context, *kmspb.AsymmetricSignRequest, ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error)
	CreateCryptoKey(context.Context, *kmspb.CreateCryptoKeyRequest, ...gax.CallOption) (*kmspb.CryptoKey, error)
	Close() error
}

// KmsClientFactory override for unittest
var KmsClientFactory = func() (KmsClient, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kms client: %v", err)
	}

	return client, nil
}

// Provider implements Provider interface for KMS
type Provider struct {
	KmsClient

	tc       cryptoprov.TokenConfig
	endpoint string
	keyring  string
}

// Init configures Kms based hsm impl
func Init(tc cryptoprov.TokenConfig) (*Provider, error) {
	kmsAttributes := parseKmsAttributes(tc.Attributes())
	endpoint := kmsAttributes["Endpoint"]
	keyring := kmsAttributes["Keyring"]

	p := &Provider{
		endpoint: endpoint,
		keyring:  keyring,
		tc:       tc,
	}

	var err error
	p.KmsClient, err = KmsClientFactory()
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create KMS client")
	}

	return p, nil
}

func parseKmsAttributes(attributes string) map[string]string {
	var kmsAttributes = make(map[string]string)

	attrs := strings.Split(attributes, ",")
	for _, v := range attrs {
		kmsAttr := strings.Split(v, "=")
		kmsAttributes[strings.TrimSpace(kmsAttr[0])] = strings.TrimSpace(kmsAttr[1])
	}

	return kmsAttributes
}

// Manufacturer returns manufacturer for the provider
func (p *Provider) Manufacturer() string {
	return p.tc.Manufacturer()
}

// Model returns model for the provider
func (p *Provider) Model() string {
	return p.tc.Model()
}

// CurrentSlotID returns current slot id. For KMS only one slot is assumed to be available.
func (p *Provider) CurrentSlotID() uint {
	return 0
}

// GenerateRSAKey creates signer using randomly generated RSA key
func (p *Provider) GenerateRSAKey(label string, bits int, purpose int) (crypto.PrivateKey, error) {
	ctx := context.Background()

	pbpurpose := kmspb.CryptoKey_ASYMMETRIC_SIGN
	if purpose == 2 {
		pbpurpose = kmspb.CryptoKey_ASYMMETRIC_DECRYPT
	}

	algorithm := kmspb.CryptoKeyVersion_RSA_SIGN_PKCS1_2048_SHA256
	switch bits {
	case 2048:
		algorithm = kmspb.CryptoKeyVersion_RSA_SIGN_PKCS1_2048_SHA256
	case 3072:
		algorithm = kmspb.CryptoKeyVersion_RSA_SIGN_PKCS1_3072_SHA256
	case 4096:
		algorithm = kmspb.CryptoKeyVersion_RSA_SIGN_PKCS1_4096_SHA512
	default:
		return nil, errors.Errorf("unsupported key size: %d", bits)
	}

	req := &kmspb.CreateCryptoKeyRequest{
		Parent:      p.keyring,
		CryptoKeyId: guid.MustCreate(),
		CryptoKey: &kmspb.CryptoKey{
			Purpose: pbpurpose,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: algorithm,
			},
			Labels: map[string]string{
				"label": label,
			},
		},
	}
	return p.genKey(ctx, req, label)
}

func (p *Provider) genKey(ctx context.Context, req *kmspb.CreateCryptoKeyRequest, label string) (crypto.PrivateKey, error) {
	resp, err := p.KmsClient.CreateCryptoKey(ctx, req)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create key")
	}

	logger.KV(xlog.NOTICE,
		"keyID", resp.Name,
		"label", label,
	)

	// Retrieve public key from KMS
	pubKeyResp, err := p.KmsClient.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: resp.Name + "/cryptoKeyVersions/1"})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to get public key")
	}

	pub, err := parseKeyFromPEM([]byte(pubKeyResp.Pem))
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse public key")
	}
	signer := NewSigner(path.Base(resp.Name), label, pub, p)

	return signer, nil
}

func parseKeyFromPEM(bytes []byte) (interface{}, error) {
	block, _ := pem.Decode(bytes)
	if block == nil || block.Type != "PUBLIC KEY" || len(block.Headers) != 0 {
		return nil, errors.Errorf("invalid block type")
	}

	k, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return k, nil
}

// GenerateECDSAKey creates signer using randomly generated ECDSA key
func (p *Provider) GenerateECDSAKey(label string, curve elliptic.Curve) (crypto.PrivateKey, error) {
	ctx := context.Background()

	pbpurpose := kmspb.CryptoKey_ASYMMETRIC_SIGN
	algorithm := kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256

	switch curve {
	case elliptic.P256():
		algorithm = kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256
	case elliptic.P384():
		algorithm = kmspb.CryptoKeyVersion_EC_SIGN_P384_SHA384
	//case elliptic.P521():
	//  algorithm = CryptoKeyVersion_EC_SIGN_P521_SHA512
	default:
		return nil, errors.New("unsupported curve")
	}

	req := &kmspb.CreateCryptoKeyRequest{
		Parent:      p.keyring,
		CryptoKeyId: guid.MustCreate(),
		CryptoKey: &kmspb.CryptoKey{
			Purpose: pbpurpose,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: algorithm,
			},
			Labels: map[string]string{
				"label": label,
			},
		},
	}
	return p.genKey(ctx, req, label)
}

// IdentifyKey returns key id and label for the given private key
func (p *Provider) IdentifyKey(priv crypto.PrivateKey) (keyID, label string, err error) {
	if s, ok := priv.(*Signer); ok {
		return s.KeyID(), s.Label(), nil
	}

	return "", "", errors.New("not supported key")
}

// GetKey returns PrivateKey
func (p *Provider) GetKey(keyID string) (crypto.PrivateKey, error) {
	logger.Infof("keyID=" + keyID)

	ctx := context.Background()
	name := p.keyName(keyID)
	key, err := p.KmsClient.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{Name: name})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to get key")
	}

	pubResponse, err := p.KmsClient.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: name + "/cryptoKeyVersions/1"})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to get public key")
	}

	pub, err := parseKeyFromPEM([]byte(pubResponse.Pem))
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse public key")
	}
	signer := NewSigner(keyID, key.Labels["label"], pub, p)
	return signer, nil
}

// EnumTokens lists tokens. For KMS currentSlotOnly is ignored and only one slot is assumed to be available.
func (p *Provider) EnumTokens(currentSlotOnly bool, slotInfoFunc func(slotID uint, description, label, manufacturer, model, serial string) error) error {
	return slotInfoFunc(p.CurrentSlotID(),
		"",
		"",
		p.Manufacturer(),
		p.Model(),
		"")
}

// EnumKeys returns list of keys on the slot. For KMS slotID is ignored.
func (p *Provider) EnumKeys(slotID uint, prefix string, keyInfoFunc func(id, label, typ, class, currentVersionID string, creationTime *time.Time) error) error {
	logger.Tracef("host=%s, slotID=%d, prefix=%q", p.endpoint, slotID, prefix)

	iter := p.KmsClient.ListCryptoKeys(
		context.Background(),
		&kmspb.ListCryptoKeysRequest{
			Parent: p.keyring,
		},
	)

	for {
		key, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Trace(err)
		}

		if key.Primary != nil && key.Primary.State != kmspb.CryptoKeyVersion_ENABLED {
			continue
		}

		createdAt := key.CreateTime.AsTime()
		err = keyInfoFunc(
			path.Base(key.Name),
			keyLabel(key),
			key.Purpose.String(),
			key.VersionTemplate.Algorithm.String(),
			"1",
			&createdAt,
		)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (p *Provider) keyName(keyID string) string {
	return p.keyring + "/cryptoKeys/" + keyID
}

func (p *Provider) keyVersionName(keyID string) string {
	return p.keyring + "/cryptoKeys/" + keyID + "/cryptoKeyVersions/1"
}

func keyLabel(key *kmspb.CryptoKey) string {
	label := "protection=" + key.VersionTemplate.ProtectionLevel.String()
	for k, v := range key.Labels {
		if label != "" {
			label += ","
		}
		label += k + "=" + v
	}
	return label
}

// DestroyKeyPairOnSlot destroys key pair on slot. For KMS slotID is ignored and KMS retire API is used to destroy the key.
func (p *Provider) DestroyKeyPairOnSlot(slotID uint, keyID string) error {
	resp, err := p.KmsClient.DestroyCryptoKeyVersion(context.Background(),
		&kmspb.DestroyCryptoKeyVersionRequest{
			Name: p.keyVersionName(keyID),
		})
	if err != nil {
		return errors.Annotatef(err, "failed to schedule key deletion: %s", keyID)
	}
	logger.Noticef("id=%s, deletion_time=%v",
		keyID, resp.DestroyTime.AsTime().Format(time.RFC3339))

	return nil
}

// KeyInfo retrieves info about key with the specified id
func (p *Provider) KeyInfo(slotID uint, keyID string, includePublic bool, keyInfoFunc func(id, label, typ, class, currentVersionID, pubKey string, creationTime *time.Time) error) error {

	ctx := context.Background()
	name := p.keyName(keyID)

	logger.KV(xlog.DEBUG, "key", name)

	key, err := p.KmsClient.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{Name: name})
	if err != nil {
		return errors.Annotatef(err, "failed to describe key, id=%s", keyID)
	}

	pubKey := ""
	if includePublic {
		pub, err := p.KmsClient.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: name + "/cryptoKeyVersions/1"})
		if err != nil {
			return errors.Annotatef(err, "failed to get public key, id=%s", keyID)
		}
		pubKey = pub.Pem
	}

	createdAt := key.CreateTime.AsTime()
	err = keyInfoFunc(
		path.Base(key.Name),
		keyLabel(key),
		key.Purpose.String(),
		key.VersionTemplate.Algorithm.String(),
		"1",
		pubKey,
		&createdAt,
	)
	if err != nil {
		return errors.Trace(err)
	}
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ExportKey returns PKCS#11 URI for specified key ID.
// It does not return key bytes
func (p *Provider) ExportKey(keyID string) (string, []byte, error) {
	uri := fmt.Sprintf("pkcs11:manufacturer=%s;id=%s;serial=1;type=private",
		ProviderName,
		keyID,
	)

	return uri, []byte(uri), nil
}

// FindKeyPairOnSlot retrieves a previously created asymmetric key, using a specified slot.
func (p *Provider) FindKeyPairOnSlot(slotID uint, keyID, label string) (crypto.PrivateKey, error) {
	return nil, errors.Errorf("unsupported command for this crypto provider")
}

// Close allocated resources and file reloader
func (p *Provider) Close() error {
	if p.KmsClient != nil {
		p.KmsClient.Close()
		p.KmsClient = nil
	}
	return nil
}

// KmsLoader provides loader for KMS provider
func KmsLoader(tc cryptoprov.TokenConfig) (cryptoprov.Provider, error) {
	p, err := Init(tc)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return p, nil
}
