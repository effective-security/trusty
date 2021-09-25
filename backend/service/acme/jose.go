package acme

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/juju/errors"
	acmemodel "github.com/martinisecurity/trusty/acme/model"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/martinisecurity/trusty/internal/db"
	"gopkg.in/square/go-jose.v2"
)

var errSigAlg = errors.New("no signature algorithms suitable for given key type")

// jwsAuthType represents whether a given POST request is authenticated using
// a JWS with an embedded JWK (v1 ACME style, new-account, revoke-cert) or an
// embeded Key ID (v2 AMCE style) or an unsupported/unknown auth type.
type jwsAuthType int

const (
	embeddedJWK jwsAuthType = iota
	embeddedKeyID
	invalidAuthType
)

func sigAlgorithmForECDSAKey(key *ecdsa.PublicKey) (string, error) {
	params := key.Params()
	switch params.Name {
	case "P-256":
		return string(jose.ES256), nil
	case "P-384":
		return string(jose.ES384), nil
	case "P-521":
		return string(jose.ES512), nil
	}
	return "", errors.Trace(errSigAlg)
}

// sigAlgorithmForKey returns signature algorithm for provided key
func sigAlgorithmForKey(key crypto.PublicKey) (string, error) {
	switch k := key.(type) {
	case *rsa.PublicKey:
		return string(jose.RS256), nil
	case *ecdsa.PublicKey:
		return sigAlgorithmForECDSAKey(k)
	}
	return "", errors.Trace(errSigAlg)
}

// Check that (1) there is a suitable algorithm for the provided key based on its
// Golang type, (2) the Algorithm field on the JWK is either absent, or matches
// that algorithm, and (3) the Algorithm field on the JWK is present and matches
// that algorithm. Precondition: parsedJws must have exactly one signature on
// it.
func checkAlgorithm(key *jose.JSONWebKey, parsedJWS *jose.JSONWebSignature) error {
	algorithm, err := sigAlgorithmForKey(key.Key)
	if err != nil {
		return errors.Trace(err)
	}
	jwsAlgorithm := parsedJWS.Signatures[0].Header.Algorithm
	if jwsAlgorithm != algorithm {
		return errors.Errorf(
			"signature type %q in JWS header is not supported, expected one of RS256, ES256, ES384 or ES512",
			jwsAlgorithm,
		)
	}
	if key.Algorithm != "" && key.Algorithm != algorithm {
		return errors.Errorf("algorithm %q on JWK is unacceptable", key.Algorithm)
	}
	return nil
}

// checkJWSAuthType examines a JWS' protected headers to determine if
// the request being authenticated by the JWS is identified using an embedded
// JWK or an embedded key ID. If no signatures are present, or mutually
// exclusive authentication types are specified at the same time, a problem is
// returned. checkJWSAuthType is separate from enforceJWSAuthType so that
// endpoints that need to handle both embedded JWK and embedded key ID requests
// can determine which type of request they have and act accordingly (e.g.
// acme v2 cert revocation).
func checkJWSAuthType(jws *jose.JSONWebSignature) (jwsAuthType, error) {
	// checkJWSAuthType is called after parseJWS() which defends against the
	// incorrect number of signatures.
	header := jws.Signatures[0].Header
	// There must not be a Key ID *and* an embedded JWK
	if header.KeyID != "" && header.JSONWebKey != nil {
		return invalidAuthType, v2acme.MalformedError("jwk and kid header fields are mutually exclusive")
	} else if header.KeyID != "" {
		return embeddedKeyID, nil
	} else if header.JSONWebKey != nil {
		return embeddedJWK, nil
	}
	return invalidAuthType, nil
}

// enforceJWSAuthType enforces a provided JWS has the provided auth type. If there
// is an error determining the auth type or if it is not the expected auth type
// then a problem is returned.
func enforceJWSAuthType(jws *jose.JSONWebSignature, expectedAuthType jwsAuthType) error {
	// Check the auth type for the provided JWS
	authType, err := checkJWSAuthType(jws)
	if err != nil {
		return errors.Trace(err)
	}
	// If the auth type isn't the one expected return a sensible problem based on
	// what was expected
	if authType != expectedAuthType {
		switch expectedAuthType {
		case embeddedKeyID:
			return v2acme.MalformedError("No Key ID in JWS header")
		case embeddedJWK:
			return v2acme.MalformedError("No embedded JWK in JWS header")
		}
	}
	return nil
}

// validPOSTRequest checks a *http.Request to ensure it has the headers
// a well-formed ACME POST request has, and to ensure there is a body to
// process.
func validPOSTRequest(request *http.Request) error {
	// All POSTs should have an accompanying Content-Length header
	if _, present := request.Header[header.ContentLength]; !present {
		return v2acme.ContentLengthRequiredError()
	}

	//if features.Enabled(features.EnforceV2ContentType) {
	// Per 6.2 ALL POSTs should have the correct JWS Content-Type for flattened
	// JSON serialization.
	if _, present := request.Header[header.ContentType]; !present {
		return v2acme.InvalidContentTypeError("No Content-Type header on POST. Content-Type must be %q",
			header.ApplicationJoseJSON)
	}
	if contentType := request.Header.Get(header.ContentType); contentType != header.ApplicationJoseJSON {
		return v2acme.InvalidContentTypeError("Invalid Content-Type header on POST. Content-Type must be %q",
			header.ApplicationJoseJSON)
	}
	//}

	// Per 6.4.1 "Replay-Nonce" clients should not send a Replay-Nonce header in
	// the HTTP request, it needs to be part of the signed JWS request body
	if _, present := request.Header[header.ReplayNonce]; present {
		return v2acme.MalformedError("HTTP requests should NOT contain Replay-Nonce header. Use JWS nonce field")
	}

	// All POSTs should have a non-nil body
	if request.Body == nil {
		return v2acme.MalformedError("No body on POST")
	}

	return nil
}

// validNonce checks a JWS' Nonce header to ensure it is one that the
// nonceService knows about, otherwise a bad nonce problem is returned.
// NOTE: this function assumes the JWS has already been verified with the
// correct public key.
func (s *Service) validNonce(jws *jose.JSONWebSignature) error {
	// validNonce is called after validPOSTRequest() and parseJWS() which
	// defend against the incorrect number of signatures.
	header := jws.Signatures[0].Header
	nonce := header.Nonce
	if len(nonce) == 0 {
		return v2acme.BadNonceError("JWS has no anti-replay nonce")
	}

	n, err := s.cadb.UseNonce(context.Background(), nonce)
	if err != nil || time.Now().UTC().After(n.ExpiresAt) {
		return v2acme.BadNonceError("JWS has an invalid anti-replay nonce: %q", nonce).WithSource(err)
	}
	return nil
}

// validPOSTURL checks the JWS' URL header against the expected URL based on the
// HTTP request. This prevents a JWS intended for one endpoint being replayed
// against a different endpoint. If the URL isn't present, is invalid, or
// doesn't match the HTTP request a problem is returned.
func validPOSTURL(request *http.Request, jws *jose.JSONWebSignature) error {
	// validPOSTURL is called after parseJWS() which defends against the incorrect
	// number of signatures.
	header := jws.Signatures[0].Header
	extraHeaders := header.ExtraHeaders
	// Check that there is at least one Extra Header
	if len(extraHeaders) == 0 {
		return v2acme.MalformedError("JWS header parameter 'url' required")
	}
	// Try to read a 'url' Extra Header as a string
	headerURL, ok := extraHeaders[jose.HeaderKey("url")].(string)
	if !ok || len(headerURL) == 0 {
		return v2acme.MalformedError("JWS header parameter 'url' required")
	}
	// Compute the URL we expect to be in the JWS based on the HTTP request
	expectedURL := url.URL{
		Scheme: requestProto(request),
		Host:   request.Host,
		Path:   request.RequestURI,
	}
	// Check that the URL we expect is the one that was found in the signed JWS
	// header
	if expectedURL.String() != headerURL {
		return v2acme.MalformedError("JWS header parameter 'url' incorrect. Expected %q, got %q",
			expectedURL.String(), headerURL)
	}
	return nil
}

// matchJWSURLs checks two JWS' URL headers are equal. This is used during key
// rollover to check that the inner JWS URL matches the outer JWS URL. If the
// JWS URLs do not match a problem is returned.
func matchJWSURLs(outer, inner *jose.JSONWebSignature) error {
	// Verify that the outer JWS has a non-empty URL header. This is strictly
	// defensive since the expectation is that endpoints using `matchJWSURLs`
	// have received at least one of their JWS from calling validPOSTForAccount(),
	// which checks the outer JWS has the expected URL header before processing
	// the inner JWS.
	outerURL, ok := outer.Signatures[0].Header.ExtraHeaders[jose.HeaderKey("url")].(string)
	if !ok || len(outerURL) == 0 {
		return v2acme.MalformedError("Outer JWS header parameter 'url' required")
	}

	// Verify the inner JWS has a non-empty URL header.
	innerURL, ok := inner.Signatures[0].Header.ExtraHeaders[jose.HeaderKey("url")].(string)
	if !ok || len(innerURL) == 0 {
		return v2acme.MalformedError("Inner JWS header parameter 'url' required")
	}

	// Verify that the outer URL matches the inner URL
	if outerURL != innerURL {
		return v2acme.MalformedError("Outer JWS 'url' value %q does not match inner JWS 'url' value %q",
			outerURL, innerURL)
	}

	return nil
}

// parseJWS extracts a JSONWebSignature from a byte slice. If there is an error
// reading the JWS or it is unacceptable (e.g. too many/too few signatures,
// presence of unprotected headers) a problem is returned, otherwise the parsed
// *JSONWebSignature is returned.
func parseJWS(body []byte) (*jose.JSONWebSignature, error) {
	// Parse the raw JWS JSON to check that:
	// * the unprotected Header field is not being used.
	// * the "signatures" member isn't present, just "signature".
	//
	// This must be done prior to `jose.parseSigned` since it will strip away
	// these headers.
	var unprotected struct {
		Header     map[string]string
		Signatures []interface{}
	}
	if err := json.Unmarshal(body, &unprotected); err != nil {
		return nil, v2acme.MalformedError("failed to decode JWS")
	}

	// ACME v2 never uses values from the unprotected JWS header. Reject JWS that
	// include unprotected headers.
	if unprotected.Header != nil {
		return nil, v2acme.MalformedError(
			`JWS "header" field not allowed. All headers must be in "protected" field`)
	}

	// ACME v2 never uses the "signatures" array of JSON serialized JWS, just the
	// mandatory "signature" field. Reject JWS that include the "signatures" array.
	if len(unprotected.Signatures) > 0 {
		return nil, v2acme.MalformedError(
			`JWS "signatures" field not allowed. Only the "signature" field should contain a signature`)
	}

	// Parse the JWS using go-jose and enforce that the expected one non-empty
	// signature is present in the parsed JWS.
	bodyStr := string(body)
	parsedJWS, err := jose.ParseSigned(bodyStr)
	if err != nil {
		return nil, v2acme.MalformedError("Parse error reading JWS")
	}
	if len(parsedJWS.Signatures) > 1 {
		return nil, v2acme.MalformedError("Too many signatures in POST body")
	}
	if len(parsedJWS.Signatures) == 0 {
		return nil, v2acme.MalformedError("POST JWS not signed")
	}
	if len(parsedJWS.Signatures) == 1 && len(parsedJWS.Signatures[0].Signature) == 0 {
		return nil, v2acme.MalformedError("POST JWS not signed")
	}

	return parsedJWS, nil
}

// parseJWSRequest extracts a JSONWebSignature from an HTTP POST request's body using parseJWS.
func parseJWSRequest(request *http.Request) (*jose.JSONWebSignature, error) {
	// Verify that the POST request has the expected headers
	if err := validPOSTRequest(request); err != nil {
		return nil, errors.Trace(err)
	}

	// Read the POST request body's bytes. validPOSTRequest has already checked
	// that the body is non-nil
	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, v2acme.ServerInternalError("unable to read request body").WithSource(err)
	}

	return parseJWS(bodyBytes)
}

// extractJWK extracts a JWK from a provided JWS or returns a problem.
// It expects that the JWS is using the embedded JWK style of authentication and
// does not contain an embedded Key ID. Callers should have acquired the
// provided JWS from parseJWS to ensure it has the correct number of signatures
// present.
func extractJWK(jws *jose.JSONWebSignature) (*jose.JSONWebKey, error) {
	// extractJWK expects the request to be using an embedded JWK auth type and
	// to not contain the mutually exclusive KeyID.
	if err := enforceJWSAuthType(jws, embeddedJWK); err != nil {
		return nil, errors.Trace(err)
	}

	// extractJWK must be called after parseJWS() which defends against the
	// incorrect number of signatures.
	header := jws.Signatures[0].Header
	// We can be sure that JSONWebKey is != nil because we have already called
	// enforceJWSAuthType()
	key := header.JSONWebKey

	// If the key isn't considered valid by go-jose return a problem immediately
	if !key.Valid() {
		return nil, v2acme.MalformedError("Invalid JWK in JWS header")
	}

	return key, nil
}

// validSelfAuthenticatedJWS checks that a given JWS verifies with the JWK
// embedded in the JWS itself (e.g. self-authenticated). This type of JWS
// is only used for creating new accounts or revoking a certificate by signing
// the request with the private key corresponding to the certificate's public
// key and embedding that public key in the JWS. All other request should be
// validated using `validJWSforAccount`. If the JWS validates (e.g. the JWS is
// well formed, verifies with the JWK embedded in it, the JWK meets
// policy/algorithm requirements, has the correct URL and includes a valid
// nonce) then `validSelfAuthenticatedJWS` returns the validated JWS body and
// the JWK that was embedded in the JWS. Otherwise if the valid JWS conditions
// are not met or an error occurs only a problem is returned
func (s *Service) validSelfAuthenticatedJWS(jws *jose.JSONWebSignature, request *http.Request) ([]byte, *jose.JSONWebKey, error) {
	// Extract the embedded JWK from the parsed JWS
	pubKey, err := extractJWK(jws)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	// TODO: implement KeyPolicy
	// If the key doesn't meet the GoodKey policy return a problem immediately
	//if err := keyPolicy.GoodKey(pubKey.Key); err != nil {
	//	return nil, nil, v2acme.MalformedError("weak key")
	//}

	// Verify the JWS with the embedded JWK
	payload, err := s.validJWSForKey(jws, pubKey, request)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	return payload, pubKey, nil
}

// validJWSForKey checks a provided JWS for a given HTTP request validates
// correctly using the provided JWK. If the JWS verifies the protected payload
// is returned. The key/JWS algorithms are verified and
// the JWK is checked against the keyPolicy before any signature validation is
// done. If the JWS signature validates correctly then the JWS nonce value
// and the JWS URL are verified to ensure that they are correct.
func (s *Service) validJWSForKey(
	jws *jose.JSONWebSignature,
	jwk *jose.JSONWebKey,
	request *http.Request,
) ([]byte, error) {

	// Check that the public key and JWS algorithms match expected
	if err := checkAlgorithm(jwk, jws); err != nil {
		return nil, v2acme.MalformedError("unsupported algorithm").WithSource(err)
	}

	// Verify the JWS signature with the public key.
	// NOTE: It might seem insecure for the WFE to be trusted to verify
	// client requests, i.e., that the verification should be done at the
	// RA.  However the WFE is the RA's only view of the outside world
	// *anyway*, so it could always lie about what key was used by faking
	// the signature itself.
	payload, err := jws.Verify(jwk)
	if err != nil {
		return nil, v2acme.MalformedError("JWS verification error").WithSource(err)
	}

	// Check that the JWS contains a correct Nonce header
	if err := s.validNonce(jws); err != nil {
		return nil, errors.Trace(err)
	}

	// Check that the HTTP request URL matches the URL in the signed JWS
	if err := validPOSTURL(request, jws); err != nil {
		return nil, errors.Trace(err)
	}

	// In the WFE1 package the check for the request URL required unmarshalling
	// the payload JSON to check the "resource" field of the protected JWS body.
	// This caught invalid JSON early and so we preserve this check by explicitly
	// trying to unmarshal the payload as part of the verification and failing
	// early if it isn't valid JSON.
	if len(payload) > 0 && !bytes.Equal(payload, []byte(`""`)) {
		var parsedBody struct{}
		if err := json.Unmarshal(payload, &parsedBody); err != nil {
			return nil, v2acme.MalformedError("request payload did not parse as JSON").WithSource(err)
		}
	}

	return payload, nil
}

// ValidSelfAuthenticatedPOST checks that a given POST request has a valid JWS
// using `validSelfAuthenticatedJWS`.
func (s *Service) ValidSelfAuthenticatedPOST(request *http.Request) ([]byte, *jose.JSONWebKey, error) {
	// Parse the JWS from the POST request
	jws, err := parseJWSRequest(request)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	// Extract and validate the embedded JWK from the parsed JWS
	return s.validSelfAuthenticatedJWS(jws, request)
}

// requestProto returns "http" for HTTP requests and "https" for HTTPS
// requests. It supports the use of "X-Forwarded-Proto" to override the protocol.
func requestProto(request *http.Request) string {
	proto := "http"

	// If the request was received via TLS, use `https://` for the protocol
	if request.TLS != nil {
		proto = "https"
	}

	// Allow upstream proxies  to specify the forwarded protocol. Allow this value
	// to override our own guess.
	if specifiedProto := request.Header.Get("X-Forwarded-Proto"); specifiedProto != "" {
		proto = specifiedProto
	}

	return proto
}

// ValidPOSTForAccount checks that a given POST request has a valid JWS
// using `validJWSForAccount`.
func (s *Service) ValidPOSTForAccount(
	ctx context.Context,
	request *http.Request,
) ([]byte, *jose.JSONWebSignature, *acmemodel.Registration, error) {
	// Parse the JWS from the POST request
	jws, err := parseJWSRequest(request)
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	return s.validJWSForAccount(ctx, request, jws)
}

// validJWSForAccount checks that a given JWS is valid and verifies with the
// public key associated to a known account specified by the JWS Key ID. If the
// JWS is valid (e.g. the JWS is well formed, verifies with the JWK stored for the
// specified key ID, specifies the correct URL, and has a valid nonce) then
// `validJWSForAccount` returns the validated JWS body, the parsed
// JSONWebSignature, and a pointer to the JWK's associated account. If any of
// these conditions are not met or an error occurs only a problem is returned.
func (s *Service) validJWSForAccount(
	ctx context.Context,
	request *http.Request,
	jws *jose.JSONWebSignature,
) ([]byte, *jose.JSONWebSignature, *acmemodel.Registration, error) {

	// Lookup the account and JWK for the key ID that authenticated the JWS
	pubKey, account, err := s.lookupJWK(ctx, request, jws)
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	// Verify the JWS with the JWK from the SA
	payload, err := s.validJWSForKey(jws, pubKey, request)
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	return payload, jws, account, nil
}

// lookupJWK finds a JWK associated with the Key ID present in a provided JWS,
// returning the JWK and a pointer to the associated account, or a problem.
// It expects that the JWS is using the embedded Key ID style of authentication
// and does not contain an embedded JWK. Callers should have acquired the
// provided JWS from parseJWS to ensure it has the correct number of signatures
// present.
func (s *Service) lookupJWK(
	ctx context.Context,
	request *http.Request,
	jws *jose.JSONWebSignature,
) (*jose.JSONWebKey, *acmemodel.Registration, error) {
	// We expect the request to be using an embedded Key ID auth type and to not
	// contain the mutually exclusive embedded JWK.
	if err := enforceJWSAuthType(jws, embeddedKeyID); err != nil {
		return nil, nil, errors.Trace(err)
	}

	header := jws.Signatures[0].Header
	acctID, err := db.ID(path.Base(header.KeyID))
	if err != nil {
		return nil, nil, v2acme.AccountDoesNotExistError("account %q not found", header.KeyID)
	}

	// Try to find the account for this account ID
	account, err := s.controller.GetRegistration(ctx, acctID)
	if err != nil {
		// If the account isn't found, return a suitable problem
		if db.IsNotFoundError(err) {
			return nil, nil, v2acme.AccountDoesNotExistError("account %q not found", header.KeyID)
		}

		return nil, nil, v2acme.ServerInternalError("unable to retreive account %q", header.KeyID)
	}

	// Verify the account is not deactivated
	if account.Status != v2acme.StatusValid {
		return nil, nil, v2acme.UnauthorizedError("account is not valid, has status %q", account.Status)
	}

	return account.Key, account, nil
}
