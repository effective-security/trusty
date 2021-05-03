package embed

import (
	"context"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/backend/service/ca"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/backend/service/status"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/proxy"
	"github.com/ekspand/trusty/pkg/gserver"
)

// TODO: refactor

// NewClient returns client.Client for running TrustyServer
func NewClient(s *gserver.Server) *client.Client {
	c := client.NewCtxClient(context.Background())

	if statusServiceServer, ok := s.Service(status.ServiceName).(trustypb.StatusServiceServer); ok {
		c.StatusService = client.NewStatusFromProxy(proxy.StatusServerToClient(statusServiceServer))
	}

	if authorityServiceServer, ok := s.Service(ca.ServiceName).(trustypb.AuthorityServiceServer); ok {
		c.AuthorityService = client.NewAuthorityFromProxy(proxy.AuthorityServerToClient(authorityServiceServer))
	}

	if certInfoServiceServer, ok := s.Service(cis.ServiceName).(trustypb.CertInfoServiceServer); ok {
		c.CertInfoService = client.NewCertInfoFromProxy(proxy.CertInfoServiceServerToClient(certInfoServiceServer))
	}
	return c
}
