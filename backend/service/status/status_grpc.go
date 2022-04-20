package status

import (
	"context"

	"github.com/effective-security/porto/xhttp/identity"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/internal/version"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Version returns the server version.
func (s *Service) Version(_ context.Context, _ *empty.Empty) (*pb.ServerVersion, error) {
	v := version.Current()
	return &pb.ServerVersion{
		Build:   v.Build,
		Runtime: v.Runtime,
	}, nil
}

// Server returns the server version.
func (s *Service) Server(_ context.Context, _ *empty.Empty) (*pb.ServerStatusResponse, error) {
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
func (s *Service) Caller(ctx context.Context, _ *empty.Empty) (*pb.CallerStatusResponse, error) {
	callerCtx := identity.FromContext(ctx)
	role := identity.GuestRoleName
	var id, name string
	if callerCtx != nil {
		caller := callerCtx.Identity()
		//id = caller.UserID()
		//name = caller.Name()
		role = caller.Role()
	}

	// TODO: change CallerStatusResponse
	res := &pb.CallerStatusResponse{
		Id:   id,
		Name: name,
		Role: role,
	}

	return res, nil
}
