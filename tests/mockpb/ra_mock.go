package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/gogo/protobuf/proto"
)

// MockRAServer for testing
type MockRAServer struct {
	pb.RAServiceServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockRAServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// SetError sets an error response
func (m *MockRAServer) SetError(err error) {
	m.Err = err
}

// RegisterRoot registers root CA
func (m *MockRAServer) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RootsResponse), nil
}

// RegisterCertificate registers certificate
func (m *MockRAServer) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateResponse), nil
}

// RevokeCertificate returns the revoked certificate
func (m *MockRAServer) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RevokedCertificateResponse), nil
}
