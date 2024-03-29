// Code generated by protoc-gen-go-proxy. DO NOT EDIT.
// source: cis.proto

package proxypb

import (
	"context"

	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/trusty/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type proxyCISServer struct {
	srv pb.CISServer
}

type proxyCISClient struct {
	remote   pb.CISClient
	callOpts []grpc.CallOption
}

type postproxyCISClient struct {
	client retriable.PostRequester
}

// CISServerToClient returns pb.CISClient
func CISServerToClient(srv pb.CISServer) pb.CISClient {
	return &proxyCISServer{srv}
}

// NewCISClient returns instance of the CISClient
func NewCISClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) pb.CISServer {
	return &proxyCISClient{
		remote:   pb.NewCISClient(conn),
		callOpts: callOpts,
	}
}

// NewCISClientFromProxy returns instance of CISClient
func NewCISClientFromProxy(proxy pb.CISClient) pb.CISServer {
	return &proxyCISClient{
		remote: proxy,
	}
}

// NewCISClientFromProxy returns instance of CISClient
func NewHTTPCISClient(client retriable.PostRequester) pb.CISServer {
	return &postproxyCISClient{
		client: client,
	}
}

// Roots returns the root CAs
func (s *proxyCISServer) GetRoots(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	// add corellation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.srv.GetRoots(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Roots returns the root CAs
func (s *proxyCISClient) GetRoots(ctx context.Context, req *emptypb.Empty) (*pb.RootsResponse, error) {
	// add corellation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.remote.GetRoots(ctx, req, s.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Roots returns the root CAs
func (s *postproxyCISClient) GetRoots(ctx context.Context, req *emptypb.Empty) (*pb.RootsResponse, error) {
	var res pb.RootsResponse
	path := "/pb.CIS/GetRoots"
	_, _, err := s.client.Post(ctx, path, req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetCertificate returns the certificate
func (s *proxyCISServer) GetCertificate(ctx context.Context, req *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	// add corellation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.srv.GetCertificate(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// GetCertificate returns the certificate
func (s *proxyCISClient) GetCertificate(ctx context.Context, req *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	// add corellation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.remote.GetCertificate(ctx, req, s.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// GetCertificate returns the certificate
func (s *postproxyCISClient) GetCertificate(ctx context.Context, req *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	var res pb.CertificateResponse
	path := "/pb.CIS/GetCertificate"
	_, _, err := s.client.Post(ctx, path, req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
