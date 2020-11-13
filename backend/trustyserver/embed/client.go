package embed

import (
	"context"

	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/proxy"
)

// NewClient returns client.Client for running TrustyServer
func NewClient(s *trustyserver.TrustyServer) *client.Client {
	c := client.NewCtxClient(context.Background())

	c.Status = client.NewStatusFromProxy(proxy.StatusServerToClient(s.StatusServer))
	c.Authority = client.NewAuthorityFromProxy(proxy.AuthorityServerToClient(s.AuthorityServer))
	return c
}
