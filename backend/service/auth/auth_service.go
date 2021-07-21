package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/pkg/jwt"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/google/go-github/github"
	"github.com/juju/errors"
	"golang.org/x/oauth2"
	googleapioauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// ServiceName provides the Service Name for this package
const ServiceName = "auth"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "auth")

const (
	evtTokenIssued    = "token_issued"
	evtTokenRefreshed = "token_refreshed"
)

// Service defines the Status service
type Service struct {
	GithubBaseURL *url.URL
	GoogleBaseURL *url.URL

	server    *gserver.Server
	cfg       *config.Configuration
	oauthProv *oauth2client.Provider
	db        orgsdb.OrgsDb
	jwt       jwt.Provider
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, oauthProv *oauth2client.Provider, sql orgsdb.OrgsDb, jwt jwt.Provider) error {
		svc := &Service{
			server:    server,
			cfg:       cfg,
			oauthProv: oauthProv,
			db:        sql,
			jwt:       jwt,
		}

		if cfg.Github.BaseURL != "" {
			u, err := url.Parse(cfg.Github.BaseURL)
			if err != nil {
				return errors.Trace(err)
			}
			svc.GithubBaseURL = u
		}

		if cfg.Google.BaseURL != "" {
			u, err := url.Parse(cfg.Google.BaseURL)
			if err != nil {
				return errors.Trace(err)
			}
			svc.GoogleBaseURL = u
		}

		server.AddService(svc)
		return nil
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return ServiceName
}

// IsReady indicates that the service is ready to serve its end-points
func (s *Service) IsReady() bool {
	return true
}

// Close the subservices and it's resources
func (s *Service) Close() {
	logger.KV(xlog.INFO, "closed", ServiceName)
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
	r.GET(v1.PathForAuthURL, s.AuthURLHandler())
	r.GET(v1.PathForAuthGithubCallback, s.GithubCallbackHandler())
	r.GET(v1.PathForAuthGoogleCallback, s.GoogleCallbackHandler())
	r.GET(v1.PathForAuthTokenRefresh, s.RefreshHandler())
	r.GET(v1.PathForAuthDone, s.AuthDoneHandler())
}

// OAuthConfig returns oauth2client.Config,
// to be used in tests
func (s *Service) OAuthConfig(provider string) *oauth2client.Config {
	return s.oauthProv.Client(provider).Config()
}

// AuthURLHandler handles v1.PathForAuthURL
func (s *Service) AuthURLHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		redirectURL, ok := r.URL.Query()["redirect_url"]
		if !ok || len(redirectURL) != 1 || redirectURL[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing redirect_url parameter"))
			return
		}

		deviceID, ok := r.URL.Query()["device_id"]
		if !ok || len(deviceID) != 1 || deviceID[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing device_id parameter"))
			return
		}

		sts := ""
		providerParam, ok := r.URL.Query()["sts"]
		if !ok || len(providerParam) != 1 || providerParam[0] == "" {
			// use github oauth2 provider by default
			sts = v1.ProviderGithub
		} else {
			sts = providerParam[0]
		}

		redirectURLCallback := ""
		switch sts {
		case v1.ProviderGithub:
			redirectURLCallback = v1.PathForAuthGithubCallback
		case v1.ProviderGoogle:
			redirectURLCallback = v1.PathForAuthGoogleCallback
		default:
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid oauth2 provider"))
			return
		}

		js, _ := json.Marshal(&authState{
			RedirectURL: redirectURL[0],
			DeviceID:    deviceID[0],
		})
		responseMode := oauth2.SetAuthURLParam("response_mode", "query")
		oauth2ResponseType := oauth2.SetAuthURLParam("response_type", "code")

		o := s.OAuthConfig(sts)
		conf := &oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			RedirectURL:  s.cfg.TrustyClient.ServerURL[config.WFEServerName][0] + redirectURLCallback,
			Scopes:       o.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  o.AuthURL,
				TokenURL: o.TokenURL,
			},
		}

		nonce := oauth2.SetAuthURLParam("nonce", certutil.RandomString(12))
		res := &v1.AuthStsURLResponse{
			// Redirect user to consent page to ask for permission
			// for the scopes specified above.
			URL: conf.AuthCodeURL(base64.RawURLEncoding.EncodeToString(js), oauth2ResponseType, responseMode, nonce),
		}

		logger.Tracef("reqRedirectURL=%q, confRedirectURL=%q, deviceID=%s, sts=%s, url=%q",
			redirectURL[0], conf.RedirectURL, deviceID[0], sts, res.URL)

		marshal.WriteJSON(w, r, res)
	}
}

// GithubCallbackHandler handles v1.PathForAuthGithubCallback
func (s *Service) GithubCallbackHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		code, ok := r.URL.Query()["code"]
		if !ok || len(code) != 1 || code[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing code parameter"))
			return
		}

		state, ok := r.URL.Query()["state"]
		if !ok || len(state) != 1 || state[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing state parameter"))
			return
		}

		js, err := base64.RawURLEncoding.DecodeString(state[0])
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid state parameter: %s", err.Error()))
			return
		}

		var oauthStatus authState
		if err = json.Unmarshal(js, &oauthStatus); err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("failed to decode state parameter: %s", err.Error()))
			return
		}

		o := s.OAuthConfig(v1.ProviderGithub)
		conf := &oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			RedirectURL:  s.cfg.TrustyClient.ServerURL[config.WFEServerName][0] + v1.PathForAuthGithubCallback,
			Scopes:       o.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  o.AuthURL,
				TokenURL: o.TokenURL,
			},
		}

		ctx := context.Background()
		token, err := conf.Exchange(ctx, code[0])
		if err != nil {
			err = errors.Trace(err)
			logger.Debugf("reason=Exchange, confRedirectURL=%q, AuthURL=%q, TokenURL=%q, sec=%q, err=[%s]",
				conf.RedirectURL, o.AuthURL, o.TokenURL, o.ClientSecret, errors.Details(err))
			marshal.WriteJSON(w, r, httperror.WithForbidden("authorization failed: %s", err.Error()).WithCause(err))
			return
		}

		logger.Debugf("redirectURL=%q, deviceID=%s, token=[%+v]",
			oauthStatus.RedirectURL, oauthStatus.DeviceID, *token)

		if !token.Valid() {
			marshal.WriteJSON(w, r, httperror.WithForbidden("retreived invalid token"))
			return
		}

		client := github.NewClient(conf.Client(ctx, token))
		if s.GithubBaseURL != nil {
			client.BaseURL = s.GithubBaseURL
		}

		ghu, _, err := client.Users.Get(ctx, "")
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("unable to retrieve user info: %s", err.Error()).WithCause(err))
			return
		}

		uemail := db.String(ghu.Email)
		if uemail == "" {
			marshal.WriteJSON(w, r, httperror.WithForbidden("please update your GitHub profile with valid email"))
			return
		}

		user := &model.User{
			ExternalID:   fmt.Sprintf("%d", ghu.GetID()),
			Provider:     v1.ProviderGithub,
			Login:        db.String(ghu.Login),
			Name:         db.String(ghu.Name),
			Email:        uemail,
			Company:      db.String(ghu.Company),
			AvatarURL:    db.String(ghu.AvatarURL),
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		}

		if !token.Expiry.IsZero() {
			user.TokenExpiresAt = db.NullTime(&token.Expiry)
		}

		user, err = s.db.LoginUser(ctx, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to login user: %s", err.Error()).WithCause(err))
			return
		}
		logger.KV(xlog.DEBUG, "user", user)

		dto := user.ToDto()
		// initial token is valid for 1 min, the client has to refresh it
		validFor := time.Minute
		if oauthStatus.DeviceID == s.server.Hostname() {
			// on the same host where the server is running on, allow for 8 hours
			validFor = 8 * 60 * time.Minute
			logger.Noticef("device=%s, email=%s, token_valid_for=%v",
				oauthStatus.DeviceID, uemail, validFor)
		}

		audience := s.server.Configuration().IdentityMap.JWT.Audience
		tokenStr, _, err := s.jwt.SignToken(dto.ID, user.Email, audience, validFor)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to sign JWT: %s", err.Error()).WithCause(err))
			return
		}

		redirect := fmt.Sprintf("%s?token=%s&device_id=%s", oauthStatus.RedirectURL, tokenStr, oauthStatus.DeviceID)
		logger.KV(xlog.DEBUG, "redirect", redirect)

		s.server.Audit(
			ServiceName,
			evtTokenIssued,
			user.Email,
			oauthStatus.DeviceID,
			0,
			fmt.Sprintf("ID=%s, ExternalID=%s, email=%s, name=%q",
				dto.ID, dto.ExternalID, dto.Email, dto.Name),
		)

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

// GoogleCallbackHandler handles v1.PathForAuthGoogleCallback
func (s *Service) GoogleCallbackHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		code, ok := r.URL.Query()["code"]
		if !ok || len(code) != 1 || code[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing code parameter"))
			return
		}

		state, ok := r.URL.Query()["state"]
		if !ok || len(state) != 1 || state[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing state parameter"))
			return
		}

		js, err := base64.RawURLEncoding.DecodeString(state[0])
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid state parameter: %s", err.Error()))
			return
		}

		var oauthStatus authState
		if err = json.Unmarshal(js, &oauthStatus); err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("failed to decode state parameter: %s", err.Error()))
			return
		}

		o := s.OAuthConfig(v1.ProviderGoogle)
		conf := &oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			RedirectURL:  s.cfg.TrustyClient.ServerURL[config.WFEServerName][0] + v1.PathForAuthGoogleCallback,
			Scopes:       o.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  o.AuthURL,
				TokenURL: o.TokenURL,
			},
		}

		ctx := context.Background()
		token, err := conf.Exchange(ctx, code[0])
		if err != nil {
			err = errors.Trace(err)
			logger.Debugf("reason=Exchange, confRedirectURL=%q, AuthURL=%q, TokenURL=%q, sec=%q, err=[%s]",
				conf.RedirectURL, o.AuthURL, o.TokenURL, o.ClientSecret, errors.Details(err))
			marshal.WriteJSON(w, r, httperror.WithForbidden("authorization failed: %s", err.Error()).WithCause(err))
			return
		}

		logger.Debugf("redirectURL=%q, deviceID=%s, token=[%+v]",
			oauthStatus.RedirectURL, oauthStatus.DeviceID, *token)

		if !token.Valid() {
			marshal.WriteJSON(w, r, httperror.WithForbidden("retreived invalid token"))
			return
		}

		opts := []option.ClientOption{option.WithTokenSource(conf.TokenSource(ctx, token))}

		if s.GoogleBaseURL != nil {
			opts = append(opts, option.WithEndpoint(s.GoogleBaseURL.Scheme+"://"+s.GoogleBaseURL.Host))
		}

		oauth2service, err := googleapioauth2.NewService(ctx, opts...)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("unable to retrieve user info: %s", err.Error()).WithCause(err))
			return
		}

		userInfo, err := oauth2service.Userinfo.V2.Me.Get().Do()
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("unable to retrieve user info: %s", err.Error()).WithCause(err))
			return
		}

		uemail := userInfo.Email
		if uemail == "" {
			marshal.WriteJSON(w, r, httperror.WithForbidden("please update your Google profile with valid email"))
			return
		}

		user := &model.User{
			ExternalID:   userInfo.Id,
			Provider:     v1.ProviderGoogle,
			Login:        db.String(&userInfo.Email),
			Name:         db.String(&userInfo.Name),
			Email:        uemail,
			Company:      "",
			AvatarURL:    db.String(&userInfo.Picture),
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		}

		if !token.Expiry.IsZero() {
			user.TokenExpiresAt = db.NullTime(&token.Expiry)
		}

		user, err = s.db.LoginUser(ctx, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to login user: %s", err.Error()).WithCause(err))
			return
		}

		logger.KV(xlog.DEBUG, "user", user)

		dto := user.ToDto()
		// initial token is valid for 1 min, the client has to refresh it
		validFor := time.Minute
		if oauthStatus.DeviceID == s.server.Hostname() {
			// on the same host where the server is running on, allow for 8 hours
			validFor = 8 * 60 * time.Minute
			logger.Noticef("device=%s, email=%s, token_valid_for=%v",
				oauthStatus.DeviceID, uemail, validFor)
		}

		audience := s.server.Configuration().IdentityMap.JWT.Audience
		tokenStr, _, err := s.jwt.SignToken(dto.ID, user.Email, audience, validFor)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to sign JWT: %s", err.Error()).WithCause(err))
			return
		}

		redirect := fmt.Sprintf("%s?token=%s&device_id=%s", oauthStatus.RedirectURL, tokenStr, oauthStatus.DeviceID)
		logger.KV(xlog.DEBUG, "redirect", redirect)

		s.server.Audit(
			ServiceName,
			evtTokenIssued,
			user.Email,
			oauthStatus.DeviceID,
			0,
			fmt.Sprintf("ID=%s, ExternalID=%s, email=%s, name=%q",
				dto.ID, dto.ExternalID, dto.Email, dto.Name),
		)

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

// RefreshHandler  for token
func (s *Service) RefreshHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		ctx := identity.FromRequest(r)
		deviceID := r.Header.Get(header.XDeviceID)
		idn := ctx.Identity()

		userID, err := db.ID(idn.UserID())
		if err != nil || userID == 0 {
			marshal.WriteJSON(w, r, httperror.WithForbidden("invalid used"))
			return
		}

		user, err := s.db.GetUser(context.Background(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		if user.Email != idn.Name() {
			marshal.WriteJSON(w, r, httperror.WithForbidden("email in the token %s does not match to registered %s", idn.Name(), user.Email))
			return
		}

		dto := user.ToDto()
		audience := s.server.Configuration().IdentityMap.JWT.Audience
		auth, claims, err := s.jwt.SignToken(idn.UserID(), user.Email, audience, 8*60*time.Minute)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to sign JWT: %s", err.Error()).WithCause(err))
			return
		}

		s.server.Audit(
			ServiceName,
			evtTokenRefreshed,
			user.Email,
			deviceID,
			0,
			fmt.Sprintf("ID=%s, ExternalID=%s, email=%s, name=%q",
				dto.ID, dto.ExternalID, dto.Email, dto.Name),
		)

		res := &v1.AuthTokenRefreshResponse{
			Authorization: &v1.Authorization{
				Version:  "v1.0",
				DeviceID: deviceID,
				UserID:   dto.ID,
				Login:    user.Login,
				Name:     user.Name,
				Email:    user.Email,
				//Role
				ExpiresAt:   time.Unix(claims.ExpiresAt, 0),
				IssuedAt:    time.Now(),
				TokenType:   "jwt",
				AccessToken: auth,
			},
			Profile: dto,
		}

		marshal.WriteJSON(w, r, res)
	}
}

// AuthDoneHandler handles v1.PathForAuthDone
func (s *Service) AuthDoneHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		token, ok := r.URL.Query()["token"]
		if !ok || len(token) != 1 || token[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing token parameter"))
			return
		}

		w.Header().Set(header.ContentType, header.TextPlain)
		fmt.Fprintf(w, "Authenticated!\n\nexport TRUSTY_AUTH_TOKEN=%s\n", token[0])
	}
}

// authState is OAuth state provided by an authenticating client
type authState struct {
	RedirectURL string `json:"redirect_url"`
	DeviceID    string `json:"device_id"`
}
