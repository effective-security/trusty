package status

import (
	"context"

	pb "github.com/go-phorce/trusty/api/v1/serverpb"
)

// Version returns the server version.
func (s *Service) Version(context.Context, *pb.EmptyRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{Version: s.server.Version()}, nil
}

// Server returns the server version.
func (s *Service) Server(context.Context, *pb.EmptyRequest) (*pb.ServerStatusResponse, error) {
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
func (s *Service) Caller(context.Context, *pb.EmptyRequest) (*pb.CallerStatusResponse, error) {
	res := &pb.CallerStatusResponse{
		Role: "guest",
	}
	return res, nil
}
