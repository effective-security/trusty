package v1

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server-side error
var (
	ErrGRPCTimeout          = status.New(codes.Unavailable, "trustyserver: request timed out").Err()
	ErrGRPCPermissionDenied = status.New(codes.PermissionDenied, "trustyserver: permission denied").Err()
	ErrGRPCInvalidArgument  = status.New(codes.InvalidArgument, "trustyserver: invalid argument").Err()

	errStringToError = map[string]error{
		ErrorDesc(ErrGRPCTimeout):          ErrGRPCTimeout,
		ErrorDesc(ErrGRPCPermissionDenied): ErrGRPCPermissionDenied,
		ErrorDesc(ErrGRPCInvalidArgument):  ErrGRPCInvalidArgument,
	}
)

// client-side error
var (
	ErrTimeout          = Error(ErrGRPCTimeout)
	ErrPermissionDenied = Error(ErrGRPCPermissionDenied)
	ErrInvalidArgument  = Error(ErrGRPCInvalidArgument)
)

// TrustyError defines gRPC server errors.
// (https://github.com/grpc/grpc-go/blob/master/rpc_util.go#L319-L323)
type TrustyError struct {
	code codes.Code
	desc string
}

// Code returns grpc/codes.Code.
func (e TrustyError) Code() codes.Code {
	return e.code
}

func (e TrustyError) Error() string {
	return e.desc
}

// Error returns TrustyError
func Error(err error) error {
	if err == nil {
		return nil
	}
	verr, ok := errStringToError[ErrorDesc(err)]
	if !ok {
		// not gRPC error
		return err
	}
	ev, ok := status.FromError(verr)
	var desc string
	if ok {
		desc = ev.Message()
	} else {
		desc = verr.Error()
	}
	return TrustyError{code: ev.Code(), desc: desc}
}

// ErrorDesc returns error description
func ErrorDesc(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}

// NewError returns new TrustyError
func NewError(code codes.Code, msgFormat string, vals ...interface{}) error {
	return TrustyError{
		code: code,
		desc: fmt.Sprintf(msgFormat, vals...),
	}
}
