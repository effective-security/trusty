package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

// MockCAServer for testing
type MockCAServer struct {
	pb.CAServiceServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockCAServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// ProfileInfo returns the certificate profile info
func (m *MockCAServer) ProfileInfo(context.Context, *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertProfileInfo), nil
}

// SignCertificate returns the certificate
func (m *MockCAServer) SignCertificate(context.Context, *pb.SignCertificateRequest) (*pb.CertificateBundle, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateBundle), nil
}

// Issuers returns the issuing CAs
func (m *MockCAServer) Issuers(context.Context, *empty.Empty) (*pb.IssuersInfoResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.IssuersInfoResponse), nil
}
