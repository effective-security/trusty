package status

import (
	"context"

	"github.com/go-phorce/dolly/xhttp/identity"
	pb "github.com/go-phorce/trusty/api/v1/serverpb"
)

// Version returns the server version.
func (s *Service) Version(_ context.Context, _ *pb.EmptyRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{Version: s.server.Version()}, nil
}

// Server returns the server version.
func (s *Service) Server(_ context.Context, _ *pb.EmptyRequest) (*pb.ServerStatusResponse, error) {
	res := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name:       s.server.Name(),
			Hostname:   s.server.Hostname(),
			ListenURLs: s.server.ListenURLs(),
			StartedAt:  s.server.StartedAt().Unix(),
			Version:    s.server.Version(),
		},
	}
	return res, nil
}

// Caller returns the status of the caller.
func (s *Service) Caller(ctx context.Context, _ *pb.EmptyRequest) (*pb.CallerStatusResponse, error) {
	callerCtx := identity.FromContext(ctx)
	role := "guest"
	if callerCtx != nil {
		role = callerCtx.Identity().Role()
	}

	res := &pb.CallerStatusResponse{
		Role: role,
	}

	return res, nil
}
