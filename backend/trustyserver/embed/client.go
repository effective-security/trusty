package embed

import (
	"context"

	"github.com/go-phorce/trusty/backend/trustyserver"
	"github.com/go-phorce/trusty/client"
	"github.com/go-phorce/trusty/client/proxy"
)

// NewClient returns client.Client for running TrustyServer
func NewClient(s *trustyserver.TrustyServer) *client.Client {
	c := client.NewCtxClient(context.Background())

	c.Status = client.NewStatusFromProxy(proxy.StatusServerToClient(s.StatusServer))

	return c
}
