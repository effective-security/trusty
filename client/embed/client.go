package embed

import (
	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/service/ca"
	"github.com/effective-security/trusty/backend/service/cis"
	"github.com/effective-security/trusty/backend/service/status"
	"github.com/effective-security/trusty/client"
	"github.com/effective-security/trusty/client/embed/proxy"
)

// NewStatusClient returns embedded StatusClient for running server
func NewStatusClient(s *gserver.Server) client.StatusClient {
	if statusServiceServer, ok := s.Service(status.ServiceName).(pb.StatusServiceServer); ok {
		return client.NewStatusClientFromProxy(proxy.StatusServerToClient(statusServiceServer))
	}
	return nil
}

// NewCAClient returns embedded CAClient for running server
func NewCAClient(s *gserver.Server) client.CAClient {
	if caServiceServer, ok := s.Service(ca.ServiceName).(pb.CAServiceServer); ok {
		return client.NewCAClientFromProxy(proxy.CAServerToClient(caServiceServer))
	}
	return nil
}

// NewCIClient returns embedded CIClient for running server
func NewCIClient(s *gserver.Server) client.CIClient {
	if cisServer, ok := s.Service(cis.ServiceName).(pb.CIServiceServer); ok {
		return client.NewCIClientFromProxy(proxy.CIServiceServerToClient(cisServer))
	}
	return nil
}
