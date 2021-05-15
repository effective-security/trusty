package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/gogo/protobuf/proto"
)

// MockAuthorityServer for testing
type MockAuthorityServer struct {
	pb.AuthorityServiceServer

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
func (m *MockAuthorityServer) ProfileInfo(context.Context, *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertProfileInfo), nil
}

// SignCertificate returns the certificate
func (m *MockAuthorityServer) SignCertificate(context.Context, *pb.SignCertificateRequest) (*pb.CertificateBundle, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateBundle), nil
}

// Issuers returns the issuing CAs
func (m *MockAuthorityServer) Issuers(context.Context, *pb.EmptyRequest) (*pb.IssuersInfoResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.IssuersInfoResponse), nil
}
