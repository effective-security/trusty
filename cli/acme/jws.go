package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"

	"github.com/juju/errors"
	jose "gopkg.in/square/go-jose.v2"
)

// JWS Represents a JWS.
type JWS struct {
	privKey      crypto.PrivateKey
	kid          string // Key identifier
	nonceService jose.NonceSource
}

// NewJWS Create a new JWS.
func NewJWS(privateKey crypto.PrivateKey, kid string, nonceService jose.NonceSource) *JWS {
	return &JWS{
		privKey:      privateKey,
		nonceService: nonceService,
		kid:          kid,
	}
}

// SetKid Sets a key identifier.
func (j *JWS) SetKid(kid string) {
	j.kid = kid
}

// SignContent Signs a content with the JWS.
func (j *JWS) SignContent(url string, content []byte) (*jose.JSONWebSignature, error) {
	var alg jose.SignatureAlgorithm
	switch k := j.privKey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		if k.Curve == elliptic.P256() {
			alg = jose.ES256
		} else if k.Curve == elliptic.P384() {
			alg = jose.ES384
		}
	}

	signKey := jose.SigningKey{
		Algorithm: alg,
		Key:       jose.JSONWebKey{Key: j.privKey, KeyID: j.kid},
	}

	options := jose.SignerOptions{
		NonceSource: j.nonceService,
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"url": url,
		},
	}

	if j.kid == "" {
		options.EmbedJWK = true
	}

	signer, err := jose.NewSigner(signKey, &options)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create jose signer")
	}

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to sign content")
	}
	return signed, nil
}

// SignEABContent Signs an external account binding content with the JWS.
func (j *JWS) SignEABContent(url, kid string, hmac []byte) (*jose.JSONWebSignature, error) {
	jwk := jose.JSONWebKey{Key: j.privKey}
	jwkJSON, err := jwk.Public().MarshalJSON()
	if err != nil {
		return nil, errors.Annotatef(err, "failed to encode JWK")
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmac},
		&jose.SignerOptions{
			EmbedJWK: false,
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				"kid": kid,
				"url": url,
			},
		},
	)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create EAB jose signer")
	}

	signed, err := signer.Sign(jwkJSON)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to sign EAB content")
	}

	return signed, nil
}
