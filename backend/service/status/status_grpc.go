package status

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/xhttp/identity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Version returns the server version.
func (s *Service) Version(_ context.Context, _ *pb.EmptyRequest) (*pb.ServerVersion, error) {
	v := version.Current()
	return &pb.ServerVersion{
		Build:   v.Build,
		Runtime: v.Runtime,
	}, nil
}

// Server returns the server version.
func (s *Service) Server(_ context.Context, _ *pb.EmptyRequest) (*pb.ServerStatusResponse, error) {
	v := version.Current()
	res := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name:       s.server.Name(),
			Hostname:   s.server.Hostname(),
			ListenUrls: s.server.ListenURLs(),
			StartedAt:  timestamppb.New(s.server.StartedAt()),
		},
		Version: &pb.ServerVersion{
			Build:   v.Build,
			Runtime: v.Runtime,
		},
	}
	return res, nil
}

// Caller returns the status of the caller.
func (s *Service) Caller(ctx context.Context, _ *pb.EmptyRequest) (*pb.CallerStatusResponse, error) {
	callerCtx := identity.FromContext(ctx)
	role := identity.GuestRoleName
	var id, name string
	if callerCtx != nil {
		caller := callerCtx.Identity()
		id = caller.UserID()
		name = caller.Name()
		role = caller.Role()
	}

	res := &pb.CallerStatusResponse{
		Id:   id,
		Name: name,
		Role: role,
	}

	return res, nil
}
