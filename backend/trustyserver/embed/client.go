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

	c.StatusService = client.NewStatusFromProxy(proxy.StatusServerToClient(s.StatusServiceServer))
	c.AuthorityService = client.NewAuthorityFromProxy(proxy.AuthorityServerToClient(s.AuthorityServiceServer))
	c.CertInfoService = client.NewCertInfoFromProxy(proxy.CertInfoServiceServerToClient(s.CertInfoServiceServer))
	return c
}
