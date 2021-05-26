package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// AuthorityClient client interface
type AuthorityClient interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error)
	// SignCertificate returns the certificate
	SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateBundle, error)
	// Issuers returns the issuing CAs
	Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error)
}

type authorityClient struct {
	remote   pb.AuthorityServiceClient
	callOpts []grpc.CallOption
}

// NewAuthority returns instance of AuthorityService client
func NewAuthority(conn *grpc.ClientConn, callOpts []grpc.CallOption) AuthorityClient {
	return &authorityClient{
		remote:   RetryAuthorityClient(conn),
		callOpts: callOpts,
	}
}

// NewAuthorityFromProxy returns instance of Authority client
func NewAuthorityFromProxy(proxy pb.AuthorityServiceClient) AuthorityClient {
	return &authorityClient{
		remote: proxy,
	}
}

// ProfileInfo returns the certificate profile info
func (c *authorityClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	return c.remote.ProfileInfo(ctx, in, c.callOpts...)
}

// SignCertificate returns the certificate
func (c *authorityClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateBundle, error) {
	return c.remote.SignCertificate(ctx, in, c.callOpts...)
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

// SignCertificate returns the certificate
func (c *retryAuthorityClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateBundle, error) {
	return c.authority.SignCertificate(ctx, in, opts...)
}

// Issuers returns the issuing CAs
func (c *retryAuthorityClient) Issuers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.Issuers(ctx, in, opts...)
}
