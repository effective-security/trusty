package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	test1KeyPrivatePEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAyNWVhtYEKJR21y9xsHV+PD/bYwbXSeNuFal46xYxVfRL5mqh
a7vttvjB/vc7Xg2RvgCxHPCqoxgMPTzHrZT75LjCwIW2K/klBYN8oYvTwwmeSkAz
6ut7ZxPv+nZaT5TJhGk0NT2kh/zSpdriEJ/3vW+mqxYbbBmpvHqsa1/zx9fSuHYc
tAZJWzxzUZXykbWMWQZpEiE0J4ajj51fInEzVn7VxV+mzfMyboQjujPh7aNJxAWS
q4oQEJJDgWwSh9leyoJoPpONHxh5nEE5AjE01FkGICSxjpZsF+w8hOTI3XXohUdu
29Se26k2B0PolDSuj0GIQU6+W9TdLXSjBb2SpQIDAQABAoIBAHw58SXYV/Yp72Cn
jjFSW+U0sqWMY7rmnP91NsBjl9zNIe3C41pagm39bTIjB2vkBNR8ZRG7pDEB/QAc
Cn9Keo094+lmTArjL407ien7Ld+koW7YS8TyKADYikZo0vAK3qOy14JfQNiFAF9r
Bw61hG5/E58cK5YwQZe+YcyBK6/erM8fLrJEyw4CV49wWdq/QqmNYU1dx4OExAkl
KMfvYXpjzpvyyTnZuS4RONfHsO8+JTyJVm+lUv2x+bTce6R4W++UhQY38HakJ0x3
XRfXooRv1Bletu5OFlpXfTSGz/5gqsfemLSr5UHncsCcFMgoFBsk2t/5BVukBgC7
PnHrAjkCgYEA887PRr7zu3OnaXKxylW5U5t4LzdMQLpslVW7cLPD4Y08Rye6fF5s
O/jK1DNFXIoUB7iS30qR7HtaOnveW6H8/kTmMv/YAhLO7PAbRPCKxxcKtniEmP1x
ADH0tF2g5uHB/zeZhCo9qJiF0QaJynvSyvSyJFmY6lLvYZsAW+C+PesCgYEA0uCi
Q8rXLzLpfH2NKlLwlJTi5JjE+xjbabgja0YySwsKzSlmvYJqdnE2Xk+FHj7TCnSK
KUzQKR7+rEk5flwEAf+aCCNh3W4+Hp9MmrdAcCn8ZsKmEW/o7oDzwiAkRCmLw/ck
RSFJZpvFoxEg15riT37EjOJ4LBZ6SwedsoGA/a8CgYEA2Ve4sdGSR73/NOKZGc23
q4/B4R2DrYRDPhEySnMGoPCeFrSU6z/lbsUIU4jtQWSaHJPu4n2AfncsZUx9WeSb
OzTCnh4zOw33R4N4W8mvfXHODAJ9+kCc1tax1YRN5uTEYzb2dLqPQtfNGxygA1DF
BkaC9CKnTeTnH3TlKgK8tUcCgYB7J1lcgh+9ntwhKinBKAL8ox8HJfkUM+YgDbwR
sEM69E3wl1c7IekPFvsLhSFXEpWpq3nsuMFw4nsVHwaGtzJYAHByhEdpTDLXK21P
heoKF1sioFbgJB1C/Ohe3OqRLDpFzhXOkawOUrbPjvdBM2Erz/r11GUeSlpNazs7
vsoYXQKBgFwFM1IHmqOf8a2wEFa/a++2y/WT7ZG9nNw1W36S3P04K4lGRNRS2Y/S
snYiqxD9nL7pVqQP2Qbqbn0yD6d3G5/7r86F7Wu2pihM8g6oyMZ3qZvvRIBvKfWo
eROL1ve1vmQF3kjrMPhhK2kr6qdWnTE5XlPllVSZFQenSTzj98AO
-----END RSA PRIVATE KEY-----
`
)

// sigAlgForKey uses `signatureAlgorithmForKey` but fails immediately using the
// testing object if the sig alg is unknown.
func sigAlgForKey(t *testing.T, key interface{}) string {
	var sigAlg string
	var err error
	// Gracefully handle the case where a non-pointer public key is given where
	// sigAlgorithmForKey always wants a pointer. It may be tempting to try and do
	// `sigAlgorithmForKey(&key)` without a type switch but this produces
	// `*interface {}` and not the desired `*rsa.PublicKey` or `*ecdsa.PublicKey`.
	switch k := key.(type) {
	case rsa.PublicKey:
		sigAlg, err = sigAlgorithmForKey(&k)
	case ecdsa.PublicKey:
		sigAlg, err = sigAlgorithmForKey(&k)
	default:
		sigAlg, err = sigAlgorithmForKey(k)
	}
	assert.NoError(t, err, "Error getting signature algorithm for key %#v", key)
	return sigAlg
}

// keyAlgForKey returns a JWK key algorithm based on the provided private key.
// Only ECDSA and RSA private keys are supported.
func keyAlgForKey(t *testing.T, key interface{}) string {
	switch key.(type) {
	case *rsa.PrivateKey, rsa.PrivateKey:
		return "RSA"
	case *ecdsa.PrivateKey, ecdsa.PrivateKey:
		return "ECDSA"
	default:
		t.Fatalf("Can't figure out keyAlgForKey: %#v", key)
	}
	return ""
}

// pubKeyForKey returns the public key of an RSA/ECDSA private key provided as
// argument.
func pubKeyForKey(t *testing.T, privKey interface{}) interface{} {
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		return k.PublicKey
	case *ecdsa.PrivateKey:
		return k.PublicKey
	default:
		t.Fatalf("Unable to get public key for private key %#v", privKey)
	}

	return nil
}

// signRequestEmbed creates a JWS for a given request body with an embedded JWK
// corresponding to the private key provided. The URL and nonce extra headers
// are set based on the additional arguments. A computed JWS, the corresponding
// embedded JWK and the JWS in serialized string form are returned.
func signRequestEmbed(
	t *testing.T,
	privateKey interface{},
	url string,
	req interface{},
	nonceService jose.NonceSource) (*jose.JSONWebSignature, *jose.JSONWebKey, string) {
	// if no key is provided default to test1KeyPrivatePEM
	var publicKey interface{}
	if privateKey == nil {
		signer := loadKey(t, []byte(test1KeyPrivatePEM))
		privateKey = signer
		publicKey = signer.Public()
	} else {
		publicKey = pubKeyForKey(t, privateKey)
	}

	signerKey := jose.SigningKey{
		Key:       privateKey,
		Algorithm: jose.SignatureAlgorithm(sigAlgForKey(t, publicKey)),
	}

	opts := &jose.SignerOptions{
		NonceSource: nonceService,
		EmbedJWK:    true,
	}
	if url != "" {
		opts.ExtraHeaders = map[jose.HeaderKey]interface{}{
			"url": url,
		}
	}

	signer, err := jose.NewSigner(signerKey, opts)
	require.NoError(t, err, "Failed to make signer")

	js, err := json.Marshal(req)
	require.NoError(t, err, "Failed to encode")

	jws, err := signer.Sign(js)
	require.NoError(t, err, "Failed to sign req")

	body := jws.FullSerialize()
	parsedJWS, err := jose.ParseSigned(body)
	require.NoError(t, err, "Failed to parse generated JWS")

	return parsedJWS, parsedJWS.Signatures[0].Header.JSONWebKey, body
}

// signRequestKeyID creates a JWS for a given request body with key ID specified
// based on the ID number provided. The URL and nonce extra headers
// are set based on the additional arguments. A computed JWS, the corresponding
// embedded JWK and the JWS in serialized string form are returned.
func signRequestKeyID(
	t *testing.T,
	keyID string,
	privateKey interface{},
	url string,
	req interface{},
	nonceService jose.NonceSource) (*jose.JSONWebSignature, *jose.JSONWebKey, string) {
	// if no key is provided default to test1KeyPrivatePEM
	if privateKey == nil {
		privateKey = loadKey(t, []byte(test1KeyPrivatePEM))
	}

	jwk := &jose.JSONWebKey{
		Key:       privateKey,
		Algorithm: keyAlgForKey(t, privateKey),
		KeyID:     keyID,
	}

	signerKey := jose.SigningKey{
		Key:       jwk,
		Algorithm: jose.RS256,
	}

	opts := &jose.SignerOptions{
		NonceSource: nonceService,
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"url": url,
		},
	}

	signer, err := jose.NewSigner(signerKey, opts)
	require.NoError(t, err, "Failed to make signer")
	var js []byte = []byte(`""`)

	if req != nil {
		js, err = json.Marshal(req)
		require.NoError(t, err, "Failed to encode")
	}

	jws, err := signer.Sign(js)
	require.NoError(t, err, "Failed to sign req")

	body := jws.FullSerialize()
	parsedJWS, err := jose.ParseSigned(body)
	require.NoError(t, err, "Failed to parse generated JWS")

	return parsedJWS, jwk, body
}

func getRequestURI(path string) string {
	t := strings.Split(path, "/v2/acme")
	return "/v2/acme" + t[1]
}

func signEABContent(t *testing.T, url, kid string, hmac []byte, privateKey interface{}) string {
	// if no key is provided default to test1KeyPrivatePEM
	if privateKey == nil {
		privateKey = loadKey(t, []byte(test1KeyPrivatePEM))
	}
	jwk := jose.JSONWebKey{Key: privateKey}
	jwkJSON, err := jwk.Public().MarshalJSON()
	require.NoError(t, err, "failed to marshal key")

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
	require.NoError(t, err, "failed to create HMAC signer")

	signed, err := signer.Sign(jwkJSON)
	require.NoError(t, err, "failed to External Account Binding sign content")

	return signed.FullSerialize()
}

func signExtraHeaders(
	t *testing.T,
	headers map[jose.HeaderKey]interface{},
	nonceService jose.NonceSource) (*jose.JSONWebSignature, string) {

	privateKey := loadKey(t, []byte(test1KeyPrivatePEM))
	signerKey := jose.SigningKey{
		Key:       privateKey,
		Algorithm: jose.SignatureAlgorithm(sigAlgForKey(t, privateKey.Public())),
	}

	opts := &jose.SignerOptions{
		NonceSource:  nonceService,
		EmbedJWK:     true,
		ExtraHeaders: headers,
	}

	signer, err := jose.NewSigner(signerKey, opts)
	require.NoError(t, err, "Failed to make signer")

	jws, err := signer.Sign([]byte(""))
	require.NoError(t, err, "Failed to sign req")

	body := jws.FullSerialize()
	parsedJWS, err := jose.ParseSigned(body)
	require.NoError(t, err, "Failed to parse generated JWS")

	return parsedJWS, body
}

// loadKey loads a private key from PEM/DER-encoded data and returns
// a `crypto.Signer`.
func loadKey(t *testing.T, keyBytes []byte) crypto.Signer {
	// pem.Decode does not return an error as its 2nd arg, but instead the "rest"
	// that was leftover from parsing the PEM block. We only care if the decoded
	// PEM block was empty for this test function.
	block, _ := pem.Decode(keyBytes)
	require.NotNil(t, block, "Unable to decode private key PEM bytes")

	// Try decoding as an RSA private key
	if rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return rsaKey
	}

	// Try decoding as a PKCS8 private key
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		// Determine the key's true type and return it as a crypto.Signer
		switch k := key.(type) {
		case *rsa.PrivateKey:
			return k
		case *ecdsa.PrivateKey:
			return k
		}
	}

	// Try as an ECDSA private key
	if ecdsaKey, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return ecdsaKey
	}

	// Nothing worked! Fail hard.
	t.Fatalf("Unable to decode private key PEM bytes")
	// NOOP - the t.Fatal() call will abort before this return
	return nil
}

// signAndPost constructs a JWS signed by the given key ID, over the given
// payload, with the protected URL set to the provided signedURL. An HTTP
// request constructed to the provided path with the encoded JWS body as the
// POST body is returned.
func signAndPost(t *testing.T, signedURL string, payload interface{}, keyID string, clientKey interface{}, ns jose.NonceSource) *http.Request {
	_, _, body := signRequestKeyID(t, keyID, clientKey, signedURL, payload, ns)
	return makePostRequestWithPath(signedURL, body)
}

func Test_RejectsNone(t *testing.T) {
	noneJWSBody := `
		{
			"header": {
				"alg": "none",
				"jwk": {
					"kty": "RSA",
					"n": "vrjT",
					"e": "AQAB"
				}
			},
			"payload": "aGkK",
  		"signature": "ghTIjrhiRl2pQ09vAkUUBbF5KziJdhzOTB-okM9SPRzU8Hyj0W1H5JA1Zoc-A-LuJGNAtYYHWqMw1SeZbT0l9FHcbMPeWDaJNkHS9jz5_g_Oyol8vcrWur2GDtB2Jgw6APtZKrbuGATbrF7g41Wijk6Kk9GXDoCnlfOQOhHhsrFFcWlCPLG-03TtKD6EBBoVBhmlp8DRLs7YguWRZ6jWNaEX-1WiRntBmhLqoqQFtvZxCBw_PRuaRw_RZBd1x2_BNYqEdOmVNC43UHMSJg3y_3yrPo905ur09aUTscf-C_m4Sa4M0FuDKn3bQ_pFrtz-aCCq6rcTIyxYpDqNvHMT2Q"
		}
	`
	noneJWS, err := jose.ParseSigned(noneJWSBody)
	require.NoError(t, err)
	noneJWK := noneJWS.Signatures[0].Header.JSONWebKey

	err = checkAlgorithm(noneJWK, noneJWS)
	require.Error(t, err, "checkAlgorithm did not reject JWS with alg: \"none\"")
	assert.Equal(t, "signature type \"none\" in JWS header is not supported, expected one of RS256, ES256, ES384 or ES512", err.Error())
}

func Test_RejectsHS256(t *testing.T) {
	hs256JWSBody := `
		{
			"header": {
				"alg": "HS256",
				"jwk": {
					"kty": "RSA",
					"n": "vrjT",
					"e": "AQAB"
				}
			},
			"payload": "aGkK",
  		"signature": "ghTIjrhiRl2pQ09vAkUUBbF5KziJdhzOTB-okM9SPRzU8Hyj0W1H5JA1Zoc-A-LuJGNAtYYHWqMw1SeZbT0l9FHcbMPeWDaJNkHS9jz5_g_Oyol8vcrWur2GDtB2Jgw6APtZKrbuGATbrF7g41Wijk6Kk9GXDoCnlfOQOhHhsrFFcWlCPLG-03TtKD6EBBoVBhmlp8DRLs7YguWRZ6jWNaEX-1WiRntBmhLqoqQFtvZxCBw_PRuaRw_RZBd1x2_BNYqEdOmVNC43UHMSJg3y_3yrPo905ur09aUTscf-C_m4Sa4M0FuDKn3bQ_pFrtz-aCCq6rcTIyxYpDqNvHMT2Q"
		}
	`

	hs256JWS, err := jose.ParseSigned(hs256JWSBody)
	require.NoError(t, err)
	hs256JWK := hs256JWS.Signatures[0].Header.JSONWebKey

	err = checkAlgorithm(hs256JWK, hs256JWS)
	require.Error(t, err, "checkAlgorithm did not reject JWS with alg: \"HS256\"")
	expected := "signature type \"HS256\" in JWS header is not supported, expected one of RS256, ES256, ES384 or ES512"
	assert.Equal(t, expected, err.Error())
}

func Test_CheckAlgorithm(t *testing.T) {
	testCases := []struct {
		key         jose.JSONWebKey
		jws         jose.JSONWebSignature
		expectedErr string
	}{
		{
			jose.JSONWebKey{
				Algorithm: "HS256",
			},
			jose.JSONWebSignature{},
			"no signature algorithms suitable for given key type",
		},
		{
			jose.JSONWebKey{
				Key: &rsa.PublicKey{},
			},
			jose.JSONWebSignature{
				Signatures: []jose.Signature{
					{
						Header: jose.Header{
							Algorithm: "HS256",
						},
					},
				},
			},
			"signature type \"HS256\" in JWS header is not supported, expected one of RS256, ES256, ES384 or ES512",
		},
		{
			jose.JSONWebKey{
				Algorithm: "HS256",
				Key:       &rsa.PublicKey{},
			},
			jose.JSONWebSignature{
				Signatures: []jose.Signature{
					{
						Header: jose.Header{
							Algorithm: "HS256",
						},
					},
				},
			},
			"signature type \"HS256\" in JWS header is not supported, expected one of RS256, ES256, ES384 or ES512",
		},
		{
			jose.JSONWebKey{
				Algorithm: "HS256",
				Key:       &rsa.PublicKey{},
			},
			jose.JSONWebSignature{
				Signatures: []jose.Signature{
					{
						Header: jose.Header{
							Algorithm: "RS256",
						},
					},
				},
			},
			"algorithm \"HS256\" on JWK is unacceptable",
		},
	}
	for _, tc := range testCases {
		err := checkAlgorithm(&tc.key, &tc.jws)
		if tc.expectedErr != "" {
			assert.Equal(t, tc.expectedErr, err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestCheckAlgorithmSuccess(t *testing.T) {
	err := checkAlgorithm(&jose.JSONWebKey{
		Algorithm: "RS256",
		Key:       &rsa.PublicKey{},
	}, &jose.JSONWebSignature{
		Signatures: []jose.Signature{
			{
				Header: jose.Header{
					Algorithm: "RS256",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("RS256 key: Expected nil error, got %q", err)
	}
	err = checkAlgorithm(&jose.JSONWebKey{
		Key: &rsa.PublicKey{},
	}, &jose.JSONWebSignature{
		Signatures: []jose.Signature{
			{
				Header: jose.Header{
					Algorithm: "RS256",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("RS256 key: Expected nil error, got %q", err)
	}

	err = checkAlgorithm(&jose.JSONWebKey{
		Algorithm: "ES256",
		Key: &ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
	}, &jose.JSONWebSignature{
		Signatures: []jose.Signature{
			{
				Header: jose.Header{
					Algorithm: "ES256",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("ES256 key: Expected nil error, got %q", err)
	}

	err = checkAlgorithm(&jose.JSONWebKey{
		Key: &ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
	}, &jose.JSONWebSignature{
		Signatures: []jose.Signature{
			{
				Header: jose.Header{
					Algorithm: "ES256",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("ES256 key: Expected nil error, got %q", err)
	}
}

func TestValidPOSTRequest(t *testing.T) {
	dummyContentLength := []string{"pretty long, idk, maybe a nibble or two?"}
	joseContentType := []string{header.ApplicationJoseJSON}

	testCases := []struct {
		Name               string
		Headers            map[string][]string
		Body               *string
		HTTPStatus         int
		ProblemDetail      string
		ErrorStatType      string
		EnforceContentType bool
	}{
		// POST requests without a Content-Length should produce a problem
		{
			Name:          "POST without a Content-Length header",
			Headers:       nil,
			HTTPStatus:    http.StatusLengthRequired,
			ProblemDetail: "missing Content-Length header",
			ErrorStatType: "ContentLengthRequired",
		},
		// POST requests with a Replay-Nonce header should produce a problem
		{
			Name: "POST with a Replay-Nonce HTTP header",
			Headers: map[string][]string{
				header.ContentLength: dummyContentLength,
				header.ContentType:   joseContentType,
				header.ReplayNonce:   {"ima-misplaced-nonce"},
			},
			HTTPStatus:    http.StatusBadRequest,
			ProblemDetail: "HTTP requests should NOT contain Replay-Nonce header. Use JWS nonce field",
			ErrorStatType: "ReplayNonceOutsideJWS",
		},
		// POST requests without a body should produce a problem
		{
			Name: "POST with an empty POST body",
			Headers: map[string][]string{
				header.ContentLength: dummyContentLength,
				header.ContentType:   joseContentType,
			},
			HTTPStatus:    http.StatusBadRequest,
			ProblemDetail: "No body on POST",
			ErrorStatType: "NoPOSTBody",
		},
		{
			Name: "POST without a Content-Type header",
			Headers: map[string][]string{
				header.ContentLength: dummyContentLength,
			},
			HTTPStatus: http.StatusUnsupportedMediaType,
			ProblemDetail: fmt.Sprintf(
				"No Content-Type header on POST. Content-Type must be %q",
				header.ApplicationJoseJSON),
			ErrorStatType:      "NoContentType",
			EnforceContentType: true,
		},
		{
			Name: "POST with an invalid Content-Type header",
			Headers: map[string][]string{
				header.ContentLength: dummyContentLength,
				header.ContentType:   {"fresh.and.rare"},
			},
			HTTPStatus: http.StatusUnsupportedMediaType,
			ProblemDetail: fmt.Sprintf(
				"Invalid Content-Type header on POST. Content-Type must be %q",
				header.ApplicationJoseJSON),
			ErrorStatType:      "WrongContentType",
			EnforceContentType: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			input := &http.Request{
				Method: http.MethodPost,
				URL:    mustParseURL("/"),
				Header: tc.Headers,
			}

			err := validPOSTRequest(input)
			assert.Error(t, err, "No error returned for invalid POST")
			require.Error(t, err)
			prob := v2acme.IsProblem(err)
			require.NotNil(t, prob)
			assert.Equal(t, v2acme.MalformedProblem.String(), prob.Type.String())
			assert.Equal(t, tc.HTTPStatus, prob.HTTPStatus)
			assert.Equal(t, tc.ProblemDetail, prob.Detail)
		})
	}
}

func TestEnforceJWSAuthType(t *testing.T) {
	nonceProv := randomNonceProvider{}
	testKeyIDJWS, _, _ := signRequestKeyID(t, "1", nil, "", nil, nonceProv)
	testEmbeddedJWS, _, _ := signRequestEmbed(t, nil, "", nil, nonceProv)

	// A hand crafted JWS that has both a Key ID and an embedded JWK
	conflictJWSBody := `
{
  "header": {
    "alg": "RS256",
    "jwk": {
      "e": "AQAB",
      "kty": "RSA",
      "n": "ppbqGaMFnnq9TeMUryR6WW4Lr5WMgp46KlBXZkNaGDNQoifWt6LheeR5j9MgYkIFU7Z8Jw5-bpJzuBeEVwb-yHGh4Umwo_qKtvAJd44iLjBmhBSxq-OSe6P5hX1LGCByEZlYCyoy98zOtio8VK_XyS5VoOXqchCzBXYf32ksVUTrtH1jSlamKHGz0Q0pRKIsA2fLqkE_MD3jP6wUDD6ExMw_tKYLx21lGcK41WSrRpDH-kcZo1QdgCy2ceNzaliBX1eHmKG0-H8tY4tPQudk-oHQmWTdvUIiHO6gSKMGDZNWv6bq74VTCsRfUEAkuWhqUhgRSGzlvlZ24wjHv5Qdlw"
    }
  },
  "protected": "eyJub25jZSI6ICJibTl1WTJVIiwgInVybCI6ICJodHRwOi8vbG9jYWxob3N0L3Rlc3QiLCAia2lkIjogInRlc3RrZXkifQ",
  "payload": "Zm9v",
  "signature": "ghTIjrhiRl2pQ09vAkUUBbF5KziJdhzOTB-okM9SPRzU8Hyj0W1H5JA1Zoc-A-LuJGNAtYYHWqMw1SeZbT0l9FHcbMPeWDaJNkHS9jz5_g_Oyol8vcrWur2GDtB2Jgw6APtZKrbuGATbrF7g41Wijk6Kk9GXDoCnlfOQOhHhsrFFcWlCPLG-03TtKD6EBBoVBhmlp8DRLs7YguWRZ6jWNaEX-1WiRntBmhLqoqQFtvZxCBw_PRuaRw_RZBd1x2_BNYqEdOmVNC43UHMSJg3y_3yrPo905ur09aUTscf-C_m4Sa4M0FuDKn3bQ_pFrtz-aCCq6rcTIyxYpDqNvHMT2Q"
}
`
	conflictJWS, err := jose.ParseSigned(conflictJWSBody)
	require.NoError(t, err, "Unable to parse conflict JWS")

	testCases := []struct {
		Name             string
		JWS              *jose.JSONWebSignature
		ExpectedAuthType jwsAuthType
		ExpectedResult   *v2acme.Problem
		ErrorStatType    string
	}{
		{
			Name:             "Key ID and embedded JWS",
			JWS:              conflictJWS,
			ExpectedAuthType: invalidAuthType,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "jwk and kid header fields are mutually exclusive",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSAuthTypeInvalid",
		},
		{
			Name:             "Key ID when expected is embedded JWK",
			JWS:              testKeyIDJWS,
			ExpectedAuthType: embeddedJWK,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "No embedded JWK in JWS header",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSAuthTypeWrong",
		},
		{
			Name:             "Embedded JWK when expected is Key ID",
			JWS:              testEmbeddedJWS,
			ExpectedAuthType: embeddedKeyID,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "No Key ID in JWS header",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSAuthTypeWrong",
		},
		{
			Name:             "Key ID when expected is KeyID",
			JWS:              testKeyIDJWS,
			ExpectedAuthType: embeddedKeyID,
			ExpectedResult:   nil,
		},
		{
			Name:             "Embedded JWK when expected is embedded JWK",
			JWS:              testEmbeddedJWS,
			ExpectedAuthType: embeddedJWK,
			ExpectedResult:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			prob := enforceJWSAuthType(tc.JWS, tc.ExpectedAuthType)
			if tc.ExpectedResult == nil {
				assert.Nil(t, prob)
			} else {
				require.Error(t, prob)
				assert.Equal(t, tc.ExpectedResult.Error(), prob.Error())
			}
		})
	}
}

func TestValidNonce(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)

	// signRequestEmbed with a `nil` nonce.NonceService will result in the
	// JWS not having a protected nonce header.
	missingNonceJWS, _, _ := signRequestEmbed(t, nil, "", "", nil)

	// signRequestEmbed with a badNonceProvider will result in the JWS
	// having an invalid nonce
	invalidNonceJWS, _, _ := signRequestEmbed(t, nil, "", "", badNonceProvider{})

	goodJWS, _, _ := signRequestEmbed(t, nil, "", "", svc)

	testCases := []struct {
		Name           string
		JWS            *jose.JSONWebSignature
		ExpectedResult *v2acme.Problem
		ErrorStatType  string
	}{
		{
			Name: "No nonce in JWS",
			JWS:  missingNonceJWS,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.BadNonceProblem,
				Detail:     "JWS has no anti-replay nonce",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSMissingNonce",
		},
		{
			Name: "Invalid nonce in JWS",
			JWS:  invalidNonceJWS,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.BadNonceProblem,
				Detail:     "JWS has an invalid anti-replay nonce: \"bad-nonce\"",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSInvalidNonce",
		},
		{
			Name:           "Valid nonce in JWS",
			JWS:            goodJWS,
			ExpectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			prob := svc.validNonce(tc.JWS)
			if tc.ExpectedResult == nil {
				assert.Nil(t, prob)
			} else {
				require.Error(t, prob)
				assert.Equal(t, tc.ExpectedResult.Error(), prob.Error())
			}
		})
	}
}

func TestValidPOSTURL(t *testing.T) {
	// A JWS and HTTP request with no extra headers
	noHeadersJWS, noHeadersJWSBody := signExtraHeaders(t, nil, randomNonceProvider{})
	noHeadersRequest := makePostRequestWithPath("http://localhost/test-path", noHeadersJWSBody)

	// A JWS and HTTP request with extra headers, but no "url" extra header
	noURLHeaders := map[jose.HeaderKey]interface{}{
		"nifty": "swell",
	}
	noURLHeaderJWS, noURLHeaderJWSBody := signExtraHeaders(t, noURLHeaders, randomNonceProvider{})
	noURLHeaderRequest := makePostRequestWithPath("http://localhost/test-path", noURLHeaderJWSBody)

	// A JWS and HTTP request with a mismatched HTTP URL to JWS "url" header
	wrongURLHeaders := map[jose.HeaderKey]interface{}{
		"url": "foobar",
	}
	wrongURLHeaderJWS, wrongURLHeaderJWSBody := signExtraHeaders(t, wrongURLHeaders, randomNonceProvider{})
	wrongURLHeaderRequest := makePostRequestWithPath("http://localhost/test-path", wrongURLHeaderJWSBody)

	correctURLHeaderJWS, _, correctURLHeaderJWSBody := signRequestEmbed(t, nil, "http://localhost/test-path", "", randomNonceProvider{})
	correctURLHeaderRequest := makePostRequestWithPath("http://localhost/test-path", correctURLHeaderJWSBody)

	testCases := []struct {
		Name           string
		JWS            *jose.JSONWebSignature
		Request        *http.Request
		ExpectedResult *v2acme.Problem
		ErrorStatType  string
	}{
		{
			Name:    "No extra headers in JWS",
			JWS:     noHeadersJWS,
			Request: noHeadersRequest,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "JWS header parameter 'url' required",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSNoExtraHeaders",
		},
		{
			Name:    "No URL header in JWS",
			JWS:     noURLHeaderJWS,
			Request: noURLHeaderRequest,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "JWS header parameter 'url' required",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSMissingURL",
		},
		{
			Name:    "Wrong URL header in JWS",
			JWS:     wrongURLHeaderJWS,
			Request: wrongURLHeaderRequest,
			ExpectedResult: &v2acme.Problem{
				Type:       v2acme.MalformedProblem,
				Detail:     "JWS header parameter 'url' incorrect. Expected \"http://localhost/test-path\", got \"foobar\"",
				HTTPStatus: http.StatusBadRequest,
			},
			ErrorStatType: "JWSMismatchedURL",
		},
		{
			Name:           "Correct URL header in JWS",
			JWS:            correctURLHeaderJWS,
			Request:        correctURLHeaderRequest,
			ExpectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			prob := validPOSTURL(tc.Request, tc.JWS)
			if tc.ExpectedResult == nil {
				assert.Nil(t, prob)
			} else {
				require.Error(t, prob)
				assert.Equal(t, tc.ExpectedResult.Error(), prob.Error())
			}
		})
	}
}

func mustParseURL(s string) *url.URL {
	if u, err := url.Parse(s); err != nil {
		panic("Cannot parse URL " + s)
	} else {
		return u
	}
}

type badNonceProvider struct {
}

func (badNonceProvider) Nonce() (string, error) {
	return "bad-nonce", nil
}

type randomNonceProvider struct {
}

func (randomNonceProvider) Nonce() (string, error) {
	return certutil.RandomString(16), nil
}

// makePostRequestWithPath creates an http.Request for localhost with method
// POST, the provided body, and the correct Content-Length. The path provided
// will be parsed as a URL and used to populate the request URL and RequestURI
func makePostRequestWithPath(uRL string, body string) *http.Request {
	request := &http.Request{
		Method:     http.MethodPost,
		RemoteAddr: "1.1.1.1:8443",
		Header: map[string][]string{
			header.ContentLength: {strconv.Itoa(len(body))},
			header.ContentType:   {header.ApplicationJoseJSON},
		},
		Body: makeBody(body),
	}
	url := mustParseURL(uRL)
	request.URL = url
	request.RequestURI = url.Path
	request.Host = url.Host
	return request
}

func makeBody(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

type singleNodeCluster struct {
	nodeID string
}

// NodeID returns the ID of the node in the cluster
func (c *singleNodeCluster) NodeID() string {
	return c.nodeID
}

// NodeHostName returns the host name of specific node
func (c *singleNodeCluster) NodeHostName(id string) (string, error) {
	if c.nodeID != id {
		return "", errors.NotFoundf("node %q", id)
	}
	return "localhost", nil
}

var cluster = &singleNodeCluster{nodeID: strconv.Itoa(int(time.Now().Unix()))}
