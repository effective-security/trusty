package workflow

import (
	"net/http"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/google/go-github/github"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// SyncGithubRepos retrieves Repos and Orgs info for the user from github
func (s *Service) SyncGithubRepos(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	o := s.OAuthConfig(user.Provider)

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

	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       user.TokenExpiresAt.Time,
	}

	client := github.NewClient(conf.Client(ctx, token))
	if s.GithubBaseURL != nil {
		client.BaseURL = s.GithubBaseURL
	}

	list, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Visibility:  "all",
		Affiliation: "owner,organization_member",
	})
	if err != nil {
		return errors.Trace(err)
	}
	logger.Debugf("src=SyncGithubRepos, repos=%d", len(list))
	json, err := marshal.EncodeBytes(marshal.PrettyPrint, list)
	if err != nil {
		return errors.Annotate(err, "failed to encode")
	}
	w.Write(json)
	/*
		for _, repo := range list {

		}
	*/

	return nil
}
