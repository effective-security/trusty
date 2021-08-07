package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// CAClient client interface
type CAClient interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error)
	// SignCertificate returns the certificate
	SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateResponse, error)
	// Issuers returns the issuing CAs
	Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error)
	// PublishCrls returns published CRLs
	PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error)
}

type authorityClient struct {
	remote   pb.CAServiceClient
	callOpts []grpc.CallOption
}

// NewCAClient returns instance of CAService client
func NewCAClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) CAClient {
	return &authorityClient{
		remote:   RetryCAClient(conn),
		callOpts: callOpts,
	}
}

// NewCAClientFromProxy returns instance of Authority client
func NewCAClientFromProxy(proxy pb.CAServiceClient) CAClient {
	return &authorityClient{
		remote: proxy,
	}
}

// ProfileInfo returns the certificate profile info
func (c *authorityClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	return c.remote.ProfileInfo(ctx, in, c.callOpts...)
}

// SignCertificate returns the certificate
func (c *authorityClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.SignCertificate(ctx, in, c.callOpts...)
}

// Issuers returns the issuing CAs
func (c *authorityClient) Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error) {
	return c.remote.Issuers(ctx, emptyReq, c.callOpts...)
}

// PublishCrls returns published CRLs
func (c *authorityClient) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	return c.remote.PublishCrls(ctx, req, c.callOpts...)
}

type retryCAClient struct {
	authority pb.CAServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryCAClient implements a CAClient.
func RetryCAClient(conn *grpc.ClientConn) pb.CAServiceClient {
	return &retryCAClient{
		authority: pb.NewCAServiceClient(conn),
	}
}

// ProfileInfo returns the certificate profile info
func (c *retryCAClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return c.authority.ProfileInfo(ctx, in, opts...)
}

// SignCertificate returns the certificate
func (c *retryCAClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.SignCertificate(ctx, in, opts...)
}

// Issuers returns the issuing CAs
func (c *retryCAClient) Issuers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.Issuers(ctx, in, opts...)
}

// PublishCrls returns published CRLs
func (c *retryCAClient) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest, opts ...grpc.CallOption) (*pb.CrlsResponse, error) {
	return c.authority.PublishCrls(ctx, req, opts...)
}
