package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"google.golang.org/grpc"
)

type cisClient struct {
	remote   pb.CertInfoServiceClient
	callOpts []grpc.CallOption
}

// NewCertInfo returns instance of CertInfoService client
func NewCertInfo(conn *grpc.ClientConn, callOpts []grpc.CallOption) CertInfoService {
	return &cisClient{
		remote:   RetryCertInfoClient(conn),
		callOpts: callOpts,
	}
}

// NewCertInfoFromProxy returns instance of CertInfoService client
func NewCertInfoFromProxy(proxy pb.CertInfoServiceClient) CertInfoService {
	return &cisClient{
		remote: proxy,
	}
}

// Roots returns the root CAs
func (c *cisClient) Roots(ctx context.Context, in *pb.EmptyRequest) (*pb.RootsResponse, error) {
	return c.remote.Roots(ctx, in, c.callOpts...)
}

type retryCertInfoClient struct {
	cis pb.CertInfoServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryCertInfoClient implements a CertInfoServiceClient.
func RetryCertInfoClient(conn *grpc.ClientConn) pb.CertInfoServiceClient {
	return &retryCertInfoClient{
		cis: pb.NewCertInfoServiceClient(conn),
	}
}

// Roots returns the root CAs
func (c *retryCertInfoClient) Roots(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return c.cis.Roots(ctx, in, opts...)
}
