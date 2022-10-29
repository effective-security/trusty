package proxy

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type statusSrv2C struct {
	srv pb.StatusServiceServer
}

// StatusServerToClient returns pb.StatusClient
func StatusServerToClient(srv pb.StatusServiceServer) pb.StatusServiceClient {
	return &statusSrv2C{srv}
}

// Version returns the server version.
func (s *statusSrv2C) Version(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.ServerVersion, error) {
	res, err := s.srv.Version(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Server returns the server status.
func (s *statusSrv2C) Server(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.ServerStatusResponse, error) {
	res, err := s.srv.Server(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Caller returns the caller status.
func (s *statusSrv2C) Caller(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.CallerStatusResponse, error) {
	res, err := s.srv.Caller(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}
