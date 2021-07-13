package gcpkmscrypto

import (
	"context"
	"crypto"
	"fmt"
	"hash/crc32"
	"io"
	"reflect"

	"github.com/juju/errors"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Signer implements crypto.Signer interface
type Signer struct {
	keyID  string
	label  string
	pubKey crypto.PublicKey
	prov   *Provider
}

// NewSigner creates new signer
func NewSigner(keyID string, label string, publicKey crypto.PublicKey, prov *Provider) crypto.Signer {
	logger.Debugf("id=%s, label=%q", keyID, label)
	return &Signer{
		keyID:  keyID,
		label:  label,
		pubKey: publicKey,
		prov:   prov,
	}
}

// KeyID returns key id of the signer
func (s *Signer) KeyID() string {
	return s.keyID
}

// Label returns key label of the signer
func (s *Signer) Label() string {
	return s.label
}

// Public returns public key for the signer
func (s *Signer) Public() crypto.PublicKey {
	return s.pubKey
}

func (s *Signer) String() string {
	return fmt.Sprintf("id=%s, label=%s",
		s.KeyID(),
		s.Label(),
	)
}

// Sign implements signing operation
func (s *Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	digestCRC32C := Crc32c(digest)

	req := &kmspb.AsymmetricSignRequest{
		Name:         s.prov.keyVersionName(s.KeyID()),
		Digest:       &kmspb.Digest{},
		DigestCrc32C: wrapperspb.Int64(int64(digestCRC32C)),
	}

	switch opts.HashFunc() {
	case crypto.SHA256:
		req.Digest.Digest = &kmspb.Digest_Sha256{Sha256: digest}
	case crypto.SHA384:
		req.Digest.Digest = &kmspb.Digest_Sha384{Sha384: digest}
	case crypto.SHA512:
		req.Digest.Digest = &kmspb.Digest_Sha512{Sha512: digest}
	default:
		return nil, errors.Errorf("unsupported hash: %s", reflect.TypeOf(opts))
	}

	result, err := s.prov.AsymmetricSign(context.Background(), req)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to sign")
	}

	// Optional, but recommended: perform integrity verification on result.
	// For more details on ensuring E2E in-transit integrity to and from Cloud KMS visit:
	// https://cloud.google.com/kms/docs/data-integrity-guidelines
	if !result.VerifiedDigestCrc32C {
		return nil, errors.Errorf("request corrupted in-transit")
	}
	// if result.Name != req.Name {
	//      return fmt.Errorf("AsymmetricSign: request corrupted in-transit")
	// }
	if int64(Crc32c(result.Signature)) != result.SignatureCrc32C.Value {
		return nil, errors.Errorf("response corrupted in-transit")
	}

	return result.Signature, nil
}

// Crc32c computes digest's CRC32C.
func Crc32c(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}
