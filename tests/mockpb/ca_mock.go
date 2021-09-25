package mockpb

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/martinisecurity/trusty/api/v1/pb"
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
func (m *MockCAServer) SignCertificate(context.Context, *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateResponse), nil
}

// Issuers returns the issuing CAs
func (m *MockCAServer) Issuers(context.Context, *empty.Empty) (*pb.IssuersInfoResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.IssuersInfoResponse), nil
}

// PublishCrls returns published CRLs
func (m *MockCAServer) PublishCrls(context.Context, *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CrlsResponse), nil
}

// GetCertificate returns the certificate
func (m *MockCAServer) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificateResponse), nil
}

// RevokeCertificate returns the revoked certificate
func (m *MockCAServer) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RevokedCertificateResponse), nil
}

// ListCertificates returns stream of Certificates
func (m *MockCAServer) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.CertificatesResponse), nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (m *MockCAServer) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RevokedCertificatesResponse), nil
}
