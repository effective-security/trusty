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
	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/ekspand/trusty/pkg/roles/jwtmapper"
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
	GithubBaseURL string

	server *trustyserver.TrustyServer
	cfg    *config.Configuration
	oauth  *oauth2client.Client
	db     db.Provider
	jwt    *jwtmapper.Provider
}

// Factory returns a factory of the service
func Factory(server *trustyserver.TrustyServer) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, oauth *oauth2client.Client, db db.Provider, jwt *jwtmapper.Provider) error {
		svc := &Service{
			server: server,
			cfg:    cfg,
			oauth:  oauth,
			db:     db,
			jwt:    jwt,
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
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
	r.GET(v1.PathForAuthURL, s.AuthURLHandler())
	r.GET(v1.PathForAuthGithubCallback, s.GithubCallbackHandler())
	r.GET(v1.PathForAuthTokenRefresh, s.RefreshHandler())
}

// OAuthConfig returns oauth2client.Config,
// to be used in tests
func (s *Service) OAuthConfig() *oauth2client.Config {
	return s.oauth.Config()
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

		js, _ := json.Marshal(&v1.AuthState{
			RedirectURL: redirectURL[0],
			DeviceID:    deviceID[0],
		})
		responseMode := oauth2.SetAuthURLParam("response_mode", "query")
		oauth2ResponseType := oauth2.SetAuthURLParam("response_type", "code")

		o := s.oauth.Config()
		conf := &oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			RedirectURL:  s.cfg.TrustyClient.PublicURL + v1.PathForAuthGithubCallback,
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

		logger.Tracef("src=getGithubURL, reqRedirectURL=%q, confRedirectURL=%q, deviceID=%s, url=%q",
			redirectURL[0], conf.RedirectURL, deviceID[0], res.URL)

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

		var oauthStatus v1.AuthState
		if err = json.Unmarshal(js, &oauthStatus); err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("failed to decode state parameter: %s", err.Error()))
			return
		}

		o := s.oauth.Config()
		conf := &oauth2.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			RedirectURL:  s.cfg.TrustyClient.PublicURL + v1.PathForAuthGithubCallback,
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
			logger.Debugf("src=githubCallbackHandler, reason=Exchange, confRedirectURL=%q, AuthURL=%q, TokenURL=%q, sec=%q, err=%q",
				conf.RedirectURL, o.AuthURL, o.TokenURL, o.ClientSecret, err.Error())
			marshal.WriteJSON(w, r, httperror.WithForbidden("authorization failed: %s", err.Error()).WithCause(err))
			return
		}
		//logger.Debugf("src=githubCallbackHandler, redirectURL=%q, deviceID=%s, token=[%+v]",
		//	oauthStatus.RedirectURL, oauthStatus.DeviceID, *token)

		if !token.Valid() {
			marshal.WriteJSON(w, r, httperror.WithForbidden("retreived invalid token"))
			return
		}

		client := github.NewClient(conf.Client(ctx, token))
		if s.GithubBaseURL != "" {
			if u, err := url.Parse(s.GithubBaseURL); err == nil {
				client.BaseURL = u
			}
		}

		ghu, _, err := client.Users.Get(ctx, "")
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("unable to retrieve user info: %s", err.Error()).WithCause(err))
			return
		}

		user := &model.User{
			GithubID:  model.NullInt64(ghu.ID),
			Login:     model.String(ghu.Login),
			Name:      model.String(ghu.Name),
			Email:     model.String(ghu.Email),
			Company:   model.String(ghu.Company),
			AvatarURL: model.String(ghu.AvatarURL),
		}

		user, err = s.db.LoginUser(ctx, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to login user: %s", err.Error()).WithCause(err))
			return
		}

		// initial token is valid for 1 min, the client has to refresh it
		dto := user.ToDto()
		auth, err := s.jwt.SignToken(dto, oauthStatus.DeviceID, time.Minute)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to sign JWT: %s", err.Error()).WithCause(err))
			return
		}

		redirect := fmt.Sprintf("%s?code=%s&device_id=%s", oauthStatus.RedirectURL, auth.AccessToken, oauthStatus.DeviceID)

		s.server.Audit(
			ServiceName,
			evtTokenIssued,
			user.Email,
			oauthStatus.DeviceID,
			0,
			fmt.Sprintf("ID=%s, GithubID=%s, email=%s, name=%q",
				dto.ID, dto.GithubID, dto.Email, dto.Name),
		)

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

// RefreshHandler  for token
func (s *Service) RefreshHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		ctx := identity.ForRequest(r)
		deviceID := r.Header.Get(header.XDeviceID)
		idn := ctx.Identity()

		userInfo, ok := idn.UserInfo().(*v1.UserInfo)
		if !ok {
			marshal.WriteJSON(w, r, httperror.WithForbidden("failed to extract User Info from the token"))
			return
		}

		userID, _ := model.ID(userInfo.ID)

		user, err := s.db.GetUser(context.Background(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		if user.Email != userInfo.Email {
			marshal.WriteJSON(w, r, httperror.WithForbidden("email in the token %s does not match to registered %s", userInfo.Email, user.Email))
			return
		}

		auth, err := s.jwt.SignToken(userInfo, deviceID, 60*time.Minute)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to sign JWT: %s", err.Error()).WithCause(err))
			return
		}

		s.server.Audit(
			ServiceName,
			evtTokenRefreshed,
			userInfo.Email,
			deviceID,
			0,
			fmt.Sprintf("ID=%s, GithubID=%s, email=%s, name=%q",
				userInfo.ID, userInfo.GithubID, userInfo.Email, userInfo.Name),
		)

		res := &v1.AuthTokenRefreshResponse{
			Authorization: auth,
			Profile:       userInfo,
		}

		marshal.WriteJSON(w, r, res)
	}
}
