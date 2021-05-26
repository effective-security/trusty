package embed

import (
	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/backend/service/ca"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/backend/service/status"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed/proxy"
	"github.com/ekspand/trusty/pkg/gserver"
)

// NewStatusClient returns embedded StatusClient for running server
func NewStatusClient(s *gserver.Server) client.StatusClient {
	if statusServiceServer, ok := s.Service(status.ServiceName).(pb.StatusServiceServer); ok {
		return client.NewStatusFromProxy(proxy.StatusServerToClient(statusServiceServer))
	}
	return nil
}

// NewAuthorityClient returns embedded AuthorityClient for running server
func NewAuthorityClient(s *gserver.Server) client.AuthorityClient {
	if authorityServiceServer, ok := s.Service(ca.ServiceName).(pb.AuthorityServiceServer); ok {
		return client.NewAuthorityFromProxy(proxy.AuthorityServerToClient(authorityServiceServer))
	}
	return nil
}

// NewCertInfoClient returns embedded CertInfoClient for running server
func NewCertInfoClient(s *gserver.Server) client.CertInfoClient {
	if certInfoServiceServer, ok := s.Service(cis.ServiceName).(pb.CertInfoServiceServer); ok {
		return client.NewCertInfoFromProxy(proxy.CertInfoServiceServerToClient(certInfoServiceServer))
	}
	return nil
}
