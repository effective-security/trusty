package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"google.golang.org/grpc"
)

type authorityClient struct {
	remote   pb.AuthorityServiceClient
	callOpts []grpc.CallOption
}

// NewAuthority returns instance of AuthorityService client
func NewAuthority(conn *grpc.ClientConn, callOpts []grpc.CallOption) AuthorityService {
	return &authorityClient{
		remote:   RetryAuthorityClient(conn),
		callOpts: callOpts,
	}
}

// NewAuthorityFromProxy returns instance of Authority client
func NewAuthorityFromProxy(proxy pb.AuthorityServiceClient) AuthorityService {
	return &authorityClient{
		remote: proxy,
	}
}

// ProfileInfo returns the certificate profile info
func (c *authorityClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	return c.remote.ProfileInfo(ctx, in, c.callOpts...)
}

// CreateCertificate returns the certificate
func (c *authorityClient) CreateCertificate(ctx context.Context, in *pb.CreateCertificateRequest) (*pb.CertificateBundle, error) {
	return c.remote.CreateCertificate(ctx, in, c.callOpts...)
}

// Issuers returns the issuing CAs
func (c *authorityClient) Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error) {
	return c.remote.Issuers(ctx, emptyReq, c.callOpts...)
}

type retryAuthorityClient struct {
	authority pb.AuthorityServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryAuthorityClient implements a AuthorityClient.
func RetryAuthorityClient(conn *grpc.ClientConn) pb.AuthorityServiceClient {
	return &retryAuthorityClient{
		authority: pb.NewAuthorityServiceClient(conn),
	}
}

// ProfileInfo returns the certificate profile info
func (c *retryAuthorityClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return c.authority.ProfileInfo(ctx, in, opts...)
}

// CreateCertificate returns the certificate
func (c *retryAuthorityClient) CreateCertificate(ctx context.Context, in *pb.CreateCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateBundle, error) {
	return c.authority.CreateCertificate(ctx, in, opts...)
}

// Issuers returns the issuing CAs
func (c *retryAuthorityClient) Issuers(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.Issuers(ctx, in, opts...)
}
