package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/gogo/protobuf/proto"
)

// MockAuthorityServer for testing
type MockAuthorityServer struct {
	trustypb.AuthorityServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockAuthorityServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// ProfileInfo returns the certificate profile info
func (m *MockAuthorityServer) ProfileInfo(context.Context, *trustypb.CertProfileInfoRequest) (*trustypb.CertProfileInfo, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.CertProfileInfo), nil
}

// CreateCertificate returns the certificate
func (m *MockAuthorityServer) CreateCertificate(context.Context, *trustypb.CreateCertificateRequest) (*trustypb.CertificateBundle, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.CertificateBundle), nil
}

// Issuers returns the issuing CAs
func (m *MockAuthorityServer) Issuers(context.Context, *trustypb.EmptyRequest) (*trustypb.IssuersInfoResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.IssuersInfoResponse), nil
}
