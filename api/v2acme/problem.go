package v2acme

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"

	"github.com/juju/errors"
)

// ProblemType defines the error types in the ACME protocol
type ProblemType string

// Error types that can be used in ACME payloads
const (
	ConnectionProblem          = ProblemType("connection")
	MalformedProblem           = ProblemType("malformed")
	ServerInternalProblem      = ProblemType("serverInternal")
	TLSProblem                 = ProblemType("tls")
	UnauthorizedProblem        = ProblemType("unauthorized")
	UnknownHostProblem         = ProblemType("unknownHost")
	RateLimitedProblem         = ProblemType("rateLimited")
	BadNonceProblem            = ProblemType("badNonce")
	InvalidEmailProblem        = ProblemType("invalidEmail")
	RejectedIdentifierProblem  = ProblemType("rejectedIdentifier")
	AccountDoesNotExistProblem = ProblemType("accountDoesNotExist")
	CAAProblem                 = ProblemType("caa")
	DNSProblem                 = ProblemType("dns")
	AlreadyRevokedProblem      = ProblemType("alreadyRevoked")

	V2ErrorNS = "urn:ietf:params:acme:error:"
)

func (t ProblemType) String() string {
	return string(t)
}

// Problem objects represent problem documents
// https://tools.ietf.org/html/draft-ietf-appsawg-http-problem-00
type Problem struct {
	Type       ProblemType `json:"type,omitempty"`
	Detail     string      `json:"detail,omitempty"`
	HTTPStatus int         `json:"status,omitempty"`

	// source is the original error
	source error
}

// IsProblem returns *Problem if error is Problem,
// or nil otherwise
func IsProblem(err error) *Problem {
	if prob, ok := err.(*Problem); ok {
		return prob
	}
	if prob, ok := errors.Cause(err).(*Problem); ok {
		return prob
	}
	return nil
}

func (prob *Problem) Error() string {
	if prob == nil {
		return "nil"
	}
	return fmt.Sprintf("%s: %s", prob.Type, prob.Detail)
}

// WithSource adds the source of the problem
func (prob *Problem) WithSource(err error) *Problem {
	prob.source = err
	return prob
}

// Source returns original error
func (prob *Problem) Source() error {
	if prob.source != nil {
		return prob.source
	}
	return nil
}

// StatusCode returns HTTP status code for the Problem
func (prob *Problem) StatusCode() int {
	if prob.HTTPStatus != 0 {
		return prob.HTTPStatus
	}
	switch prob.Type {
	case
		ConnectionProblem,
		MalformedProblem,
		TLSProblem,
		UnknownHostProblem,
		BadNonceProblem,
		InvalidEmailProblem,
		RejectedIdentifierProblem,
		AccountDoesNotExistProblem:
		return http.StatusBadRequest
	case ServerInternalProblem:
		return http.StatusInternalServerError
	case
		UnauthorizedProblem,
		CAAProblem:
		return http.StatusForbidden
	case RateLimitedProblem:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// BadNonceError returns a v2acme.Problem with a BadNonceProblem
// and a 400 Bad Request status code.
func BadNonceError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       BadNonceProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// RejectedIdentifierError returns a v2acme.Problem with a RejectedIdentifierProblem
// and a 400 Bad Request status code.
func RejectedIdentifierError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       RejectedIdentifierProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ConflictError returns a v2acme.Problem with a MalformedProblem
// and a 409 Conflict status code.
func ConflictError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusConflict,
	}
}

// AlreadyRevokedError returns a v2acme.Problem with a AlreadyRevokedProblem
// and a 400 Bad Request status code.
func AlreadyRevokedError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       AlreadyRevokedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// MalformedError returns a v2acme.Problem with a MalformedProblem
// and a 400 Bad Request status code.
func MalformedError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// NotFoundError returns a v2acme.Problem with a MalformedProblem
// and a 404 Not Found status code.
func NotFoundError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusNotFound,
	}
}

// ServerInternalError returns a v2acme.Problem with a ServerInternalProblem
// and a 500 Internal Server Failure status code.
func ServerInternalError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       ServerInternalProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusInternalServerError,
	}
}

// UnauthorizedError returns a v2acme.Problem with an UnauthorizedProblem
// and a 403 Forbidden status code.
func UnauthorizedError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       UnauthorizedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusForbidden,
	}
}

// MethodNotAllowedError returns a v2acme.Problem representing a disallowed HTTP
// method error.
func MethodNotAllowedError() *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     "Method not allowed",
		HTTPStatus: http.StatusMethodNotAllowed,
	}
}

// ContentLengthRequiredError returns a v2acme.Problem representing a missing
// Content-Length header error
func ContentLengthRequiredError() *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     "missing Content-Length header",
		HTTPStatus: http.StatusLengthRequired,
	}
}

// InvalidContentTypeError returns a v2acme.Problem suitable for a missing
// ContentType header, or an incorrect ContentType header
func InvalidContentTypeError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       MalformedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusUnsupportedMediaType,
	}
}

// InvalidEmailError returns a v2acme.Problem representing an invalid email address
// error
func InvalidEmailError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       InvalidEmailProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ConnectionFailureError returns a v2acme.Problem representing a ConnectionProblem
// error
func ConnectionFailureError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       ConnectionProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// UnknownHostError returns a v2acme.Problem representing an UnknownHostProblem error
func UnknownHostError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       UnknownHostProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// RateLimitedError returns a v2acme.Problem representing a RateLimitedProblem error
func RateLimitedError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       RateLimitedProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// TLSError returns a v2acme.Problem representing a TLSProblem error
func TLSError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       TLSProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// AccountDoesNotExistError returns a v2acme.Problem representing an
// AccountDoesNotExistProblem error
func AccountDoesNotExistError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       AccountDoesNotExistProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// CAAError returns a v2acme.Problem representing a CAAProblem
func CAAError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       CAAProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusForbidden,
	}
}

// DNSError returns a v2acme.Problem representing a DNSProblem
func DNSError(detail string, a ...interface{}) *Problem {
	return &Problem{
		Type:       DNSProblem,
		Detail:     fmt.Sprintf(detail, a...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// badTLSHeader contains the string 'HTTP /' which is returned when
// we try to talk TLS to a server that only talks HTTP
var badTLSHeader = []byte{0x48, 0x54, 0x54, 0x50, 0x2f}

// ProblemFromHTTPError returns a Problem corresponding to an error
// that occurred during HTTP request.
// Specifically, it tries to unwrap known Go error types and present something a little more
// meaningful.
func ProblemFromHTTPError(err error) *Problem {
	// unwrap cause
	return fromHTTPError(errors.Cause(err))
}

func fromHTTPError(err error) *Problem {
	// net/http wraps net.OpError in a url.Error. Unwrap them.
	if urlErr, ok := err.(*url.Error); ok {
		prob := fromHTTPError(urlErr.Err)
		prob.Detail = fmt.Sprintf("fetching %s: %s", urlErr.URL, prob.Detail)
		return prob
	}

	if tlsErr, ok := err.(tls.RecordHeaderError); ok && bytes.Compare(tlsErr.RecordHeader[:], badTLSHeader) == 0 {
		return MalformedError("server only speaks HTTP, not TLS")
	}

	if netErr, ok := err.(*net.OpError); ok {
		if fmt.Sprintf("%T", netErr.Err) == "tls.alert" {
			// All the tls.alert error strings are reasonable to hand back to a
			// user. Confirmed against Go 1.8.
			return TLSError(netErr.Error())
		} else if syscallErr, ok := netErr.Err.(*os.SyscallError); ok &&
			syscallErr.Err == syscall.ECONNREFUSED {
			return ConnectionFailureError("connection refused")
		} else if syscallErr, ok := netErr.Err.(*os.SyscallError); ok &&
			syscallErr.Err == syscall.ENETUNREACH {
			return ConnectionFailureError("network unreachable")
		} else if syscallErr, ok := netErr.Err.(*os.SyscallError); ok &&
			syscallErr.Err == syscall.ECONNRESET {
			return ConnectionFailureError("connection reset by peer")
		} else if netErr.Timeout() && netErr.Op == "dial" {
			return ConnectionFailureError("timeout during connect")
		} else if netErr.Timeout() {
			return ConnectionFailureError("timeout during %s (your server may be slow or overloaded)", netErr.Op)
		}
	}
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return ConnectionFailureError("timeout after connect (your server may be slow or overloaded)")
	}

	return ConnectionFailureError(err.Error())
}

// ProblemFromError returns Problem from error or ServerInternalError.
func ProblemFromError(err error) *Problem {
	if prob := IsProblem(err); prob != nil {
		return prob
	}

	return ServerInternalError(err.Error())
}
