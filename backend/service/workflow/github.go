package workflow

import (
	"net/http"
	"sync"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/google/go-github/github"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func (s *Service) githubClient(ctx context.Context, user *model.User) *github.Client {
	o := s.OAuthConfig(user.Provider)

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
	return client
}

// SyncGithubRepos retrieves Repos and Orgs info for the user from github
func (s *Service) SyncGithubRepos(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	client := s.githubClient(ctx, user)

	list, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Visibility:  "all",
		Affiliation: "owner,organization_member",
	})
	if err != nil {
		return errors.Trace(err)
	}
	logger.Debugf("user=%s, repos=%d", user.Email, len(list))
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

// SyncGithubOrgs syncs Orgs info for the user from github
func (s *Service) SyncGithubOrgs(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	client := s.githubClient(ctx, user)

	list, _, err := client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return errors.Trace(err)
	}
	logger.Debugf("user=%s, orgs=%d", user.Email, len(list))

	res := &v1.OrgsResponse{
		Orgs: make([]v1.Organization, 0),
	}

	ch := make(chan *v1.Organization, len(list))
	wg := sync.WaitGroup{}
	for _, org := range list {
		if org.GetRole() == "admin" && org.GetState() == "active" {
			wg.Add(1)

			logger.Debugf("user=%s, login=%s, id=%d", user.Email, org.Organization.GetLogin(), org.Organization.GetID())

			go func(login string) {
				defer wg.Done()

				o, _, err := client.Organizations.Get(ctx, login)
				if err != nil {
					logger.Errorf("user=%s, orgs=%s, err=[%v]", user.Email, login, errors.Details(err))
					return
				}

				mo, err := s.db.UpdateOrg(ctx, &model.Organization{
					ExternalID:   uint64(o.GetID()),
					Provider:     v1.ProviderGithub,
					Login:        o.GetLogin(),
					AvatarURL:    o.GetAvatarURL(),
					URL:          o.GetHTMLURL(),
					Name:         o.GetName(),
					Email:        o.GetEmail(),
					BillingEmail: o.GetBillingEmail(),
					Company:      o.GetCompany(),
					Location:     o.GetLocation(),
					Type:         o.GetType(),
					CreatedAt:    o.GetCreatedAt(),
					UpdatedAt:    o.GetCreatedAt(),
				})
				if err != nil {
					logger.Errorf("user=%s, orgs=%s, err=[%v]", user.Email, login, errors.Details(err))
					return
				}
				ch <- mo.ToDto()

				_, err = s.db.AddOrgMember(ctx, mo.ID, user.ID, "admin", v1.ProviderGithub)
				if err != nil {
					logger.Errorf("user=%s, orgs=%s, err=[%v]", user.Email, login, errors.Details(err))
				}
			}(org.Organization.GetLogin())
		}
	}
	wg.Wait()
	close(ch)

	for org := range ch {
		res.Orgs = append(res.Orgs, *org)
	}
	marshal.WritePlainJSON(w, http.StatusOK, res, marshal.PrettyPrint)
	return nil
}
