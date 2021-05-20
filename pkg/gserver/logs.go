package gserver

import (
	"context"
	"reflect"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/metrics/tags"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	warnUnaryRequestLatency = 300 * time.Millisecond
)

var (
	keyForReqPerf       = []string{"grpc", "request", "perf"}
	keyForReqSuccessful = []string{"grpc", "request", "status", "successful"}
	keyForReqFailed     = []string{"grpc", "request", "status", "failed"}
)

func (s *Server) newLogUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		defer logUnaryRequestStats(ctx, info, startTime, req, resp)
		return resp, err
	}
}

func logUnaryRequestStats(ctx context.Context, info *grpc.UnaryServerInfo, startTime time.Time, req interface{}, resp interface{}) {
	duration := time.Since(startTime)
	expensiveRequest := duration > warnUnaryRequestLatency

	remote := "no_remote_client_info"
	peerInfo, ok := peer.FromContext(ctx)
	if ok {
		remote = peerInfo.Addr.String()
	}
	responseType := info.FullMethod

	var code codes.Code
	switch _resp := resp.(type) {
	case *v1.TrustyError:
		code = _resp.Code()
	case *status.Status:
		code = _resp.Code()
	case error:
		if s, ok := status.FromError(_resp); ok {
			code = s.Code()
		}
	}

	if logger.LevelAt(xlog.DEBUG) {
		logger.Debugf("req=%v, res=%s, remote=%s, duration=%v, code=%v", reflect.TypeOf(req), responseType, remote, duration, code)
	} else if expensiveRequest {
		logger.Warningf("req=%v, res=%s, remote=%s, duration=%v, code=%v", reflect.TypeOf(req), responseType, remote, duration, code)
	}

	role := identity.GuestRoleName
	idx := identity.FromContext(ctx)
	if idx != nil {
		role = idx.Identity().Role()
	}

	tags := []metrics.Tag{
		{Name: tags.Method, Value: info.FullMethod},
		{Name: tags.Role, Value: role},
		{Name: tags.Status, Value: code.String()},
	}

	metrics.MeasureSince(keyForReqPerf, startTime, tags...)

	if code == codes.OK {
		metrics.IncrCounter(keyForReqSuccessful, 1, tags...)
	} else {
		metrics.IncrCounter(keyForReqFailed, 1, tags...)
	}
}

func newStreamInterceptor(s *Server) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.Debugf("src=newStreamInterceptor, method=%s", info.FullMethod)
		return handler(srv, ss)
	}
}
