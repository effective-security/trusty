package acme

import (
	"net/http"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xpki/certutil"
)

const randomDirKeyExplanationLink = "https://community.letsencrypt.org/t/adding-random-entries-to-the-directory/33417"

const (
	// uriNonces allows to handle HEAD for all paths under /v2/acme
	uriNonces     = v2acme.BasePath + "/:verb"
	uriNewNonce   = v2acme.BasePath + "/new-nonce"
	uriNewAccount = v2acme.BasePath + "/new-account"
	uriNewOrder   = v2acme.BasePath + "/new-order"

	uriKeyChange  = v2acme.BasePath + "/key-change"
	uriRevokeCert = v2acme.BasePath + "/revoke-cert"

	// The following URIs should have accountID in URL and
	// KeyID in the payload

	uriAccountByID = v2acme.BasePath + "/account/:acct_id"

	uriOrders    = uriAccountByID + "/orders"
	uriOrderByID = uriOrders + "/:id"

	uriAuthzByID = v2acme.BasePath + "/authz/:id"
	uriChallenge = v2acme.BasePath + "/challenge/:acct_id/:authz_id/:id"
	uriCert      = v2acme.BasePath + "/cert/:acct_id/:id"

	uriIssuer       = v2acme.BasePath + "/issuer-cert"
	uriFinalizeByID = v2acme.BasePath + "/finalize/:acct_id/:id"
)

// DirectoryHandler returns directory
func (s *Service) DirectoryHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		// identity.ForRequest(r)

		cfg := s.controller.Config()

		baseURL := cfg.Service.BaseURI
		if baseURL == "" {
			baseURL = s.cfg.TrustyClient.ServerURL["wfe"][0]
		}

		directoryEndpoints := map[string]interface{}{
			"newAccount": baseURL + uriNewAccount,
			"newNonce":   baseURL + uriNewNonce,
			"revokeCert": baseURL + uriRevokeCert,
			"newOrder":   baseURL + uriNewOrder,
			"keyChange":  baseURL + uriKeyChange,
		}

		// Add a random key to the directory in order to make sure that clients don't hardcode an
		// expected set of keys. This ensures that we can properly extend the directory when we
		// need to add a new endpoint or meta element.
		directoryEndpoints[certutil.RandomString(8)] = randomDirKeyExplanationLink

		// ACME since draft-02 describes an optional "meta" directory entry.
		// The meta entry may optionally contain a "termsOfService" URI for the current ToS.
		metaMap := map[string]interface{}{}

		if cfg.Service.SubscriberAgreementURL != "" {
			metaMap["termsOfService"] = cfg.Service.SubscriberAgreementURL
		}
		// The "meta" directory entry may also include a []string of CAA identities
		if cfg.Service.DirectoryCAAIdentity != "" {
			// The specification says caaIdentities is an array of strings.
			// In practice VA only allows configuring ONE CAA identity.
			metaMap["caaIdentities"] = []string{cfg.Service.DirectoryCAAIdentity}
		}

		// The "meta" directory entry may also include a string with a website URL
		if cfg.Service.DirectoryWebsite != "" {
			metaMap["website"] = cfg.Service.DirectoryWebsite
		}
		if len(metaMap) > 0 {
			directoryEndpoints["meta"] = metaMap
		}

		marshal.WritePlainJSON(w, http.StatusOK, directoryEndpoints, marshal.PrettyPrint)
	}
}
