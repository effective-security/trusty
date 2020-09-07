package trustyserver

import (
	"context"

	pb "github.com/go-phorce/trusty/api/v1/serverpb"
)

// Status service
type Status interface {
	// Version returns the server version.
	Version(context.Context, *pb.EmptyRequest) (*pb.VersionResponse, error)
	// Server returns the server status.
	Server(context.Context, *pb.EmptyRequest) (*pb.ServerStatusResponse, error)
}

type statusService struct {
	version string
}

// Version returns the server version.
func (s *statusService) Version(context.Context, *pb.EmptyRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{Version: s.version}, nil
}

// Version returns the server version.
func (s *statusService) Server(context.Context, *pb.EmptyRequest) (*pb.ServerStatusResponse, error) {
	return &pb.ServerStatusResponse{}, nil
}
