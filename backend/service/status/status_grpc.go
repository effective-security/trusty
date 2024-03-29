package status

import (
	"context"
	"encoding/json"
	"time"

	"github.com/effective-security/porto/xhttp/identity"
	pb "github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/trusty/internal/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Version returns the server version.
func (s *Service) Version(_ context.Context, _ *emptypb.Empty) (*pb.ServerVersion, error) {
	v := version.Current()
	return &pb.ServerVersion{
		Build:   v.Build,
		Runtime: v.Runtime,
	}, nil
}

// Server returns the server version.
func (s *Service) Server(_ context.Context, _ *emptypb.Empty) (*pb.ServerStatusResponse, error) {
	v := version.Current()
	res := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name:       s.server.Name(),
			Hostname:   s.server.Hostname(),
			ListenURLs: s.server.ListenURLs(),
			StartedAt:  s.server.StartedAt().Format(time.RFC3339),
		},
		Version: &pb.ServerVersion{
			Build:   v.Build,
			Runtime: v.Runtime,
		},
	}
	return res, nil
}

// Caller returns the status of the caller.
func (s *Service) Caller(ctx context.Context, _ *emptypb.Empty) (*pb.CallerStatusResponse, error) {
	callerCtx := identity.FromContext(ctx)
	caller := callerCtx.Identity()
	var claims []byte

	cl := caller.Claims()
	if cl != nil {
		claims, _ = json.Marshal(cl)
	}

	res := &pb.CallerStatusResponse{
		Subject: caller.Subject(),
		Role:    caller.Role(),
		Claims:  claims,
	}

	return res, nil
}
