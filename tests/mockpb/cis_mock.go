package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

// MockCIServer for testing
type MockCIServer struct {
	pb.CIServiceServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockCIServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// GetRoots returns the root CAs
func (m *MockCIServer) GetRoots(context.Context, *empty.Empty) (*pb.RootsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RootsResponse), nil
}

// GetCertificate returns the certificate
func (m *MockCIServer) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateResponse), nil
}

// GetOrgCertificates returns the Org certificates
func (m *MockCIServer) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificatesResponse), nil
}

// ListCertificates returns stream of Certificates
func (m *MockCIServer) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificatesResponse), nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (m *MockCIServer) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RevokedCertificatesResponse), nil
}
