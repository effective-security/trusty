package proxy

import (
	"context"

	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"google.golang.org/grpc"
)

type statusSrv2C struct {
	srv pb.StatusServer
}

// StatusServerToClient returns pb.StatusClient
func StatusServerToClient(srv pb.StatusServer) pb.StatusClient {
	return &statusSrv2C{srv}
}

// Version returns the server version.
func (s *statusSrv2C) Version(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.VersionResponse, error) {
	return s.srv.Version(ctx, in)
}

// Server returns the server status.
func (s *statusSrv2C) Server(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.ServerStatusResponse, error) {
	return s.srv.Server(ctx, in)
}

// Caller returns the caller status.
func (s *statusSrv2C) Caller(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.CallerStatusResponse, error) {
	return s.srv.Caller(ctx, in)
}
