package v2acme_test

import (
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
)

func Test_ProblemDetails(t *testing.T) {
	pd := &v2acme.Problem{
		Type:       v2acme.MalformedProblem,
		Detail:     "Wat? o.O",
		HTTPStatus: 403,
	}
	assert.Equal(t, pd.Error(), "malformed: Wat? o.O")
}

func Test_ProblemDetailsToStatusCode(t *testing.T) {
	testCases := []struct {
		pb         *v2acme.Problem
		statusCode int
	}{
		{&v2acme.Problem{Type: v2acme.ConnectionProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: v2acme.MalformedProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: v2acme.ServerInternalProblem}, http.StatusInternalServerError},
		{&v2acme.Problem{Type: v2acme.TLSProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: v2acme.UnauthorizedProblem}, http.StatusForbidden},
		{&v2acme.Problem{Type: v2acme.UnknownHostProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: v2acme.RateLimitedProblem}, http.StatusTooManyRequests},
		{&v2acme.Problem{Type: v2acme.BadNonceProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: v2acme.InvalidEmailProblem}, http.StatusBadRequest},
		{&v2acme.Problem{Type: "foo"}, http.StatusInternalServerError},
		{&v2acme.Problem{Type: "foo", HTTPStatus: 200}, 200},
		{&v2acme.Problem{Type: v2acme.ConnectionProblem, HTTPStatus: 200}, 200},
		{&v2acme.Problem{Type: v2acme.AccountDoesNotExistProblem}, http.StatusBadRequest},
	}

	for _, c := range testCases {
		t.Run(string(c.pb.Type), func(t *testing.T) {
			p := c.pb.StatusCode()
			assert.Equal(t, c.statusCode, p)
		})
	}
}

func Test_ProblemDetailsConvenience(t *testing.T) {
	testCases := []struct {
		pb           *v2acme.Problem
		expectedType v2acme.ProblemType
		statusCode   int
		detail       string
	}{
		{
			v2acme.InvalidEmailError("invalid email detail"),
			v2acme.InvalidEmailProblem,
			http.StatusBadRequest,
			"invalid email detail"},
		{
			v2acme.ConnectionFailureError("connection failure detail"),
			v2acme.ConnectionProblem,
			http.StatusBadRequest,
			"connection failure detail"},
		{
			v2acme.MalformedError("malformed detail"),
			v2acme.MalformedProblem,
			http.StatusBadRequest,
			"malformed detail"},
		{
			v2acme.ServerInternalError("internal error detail"),
			v2acme.ServerInternalProblem,
			http.StatusInternalServerError,
			"internal error detail"},
		{
			v2acme.UnauthorizedError("unauthorized detail"),
			v2acme.UnauthorizedProblem,
			http.StatusForbidden,
			"unauthorized detail"},
		{
			v2acme.UnknownHostError("unknown host detail"),
			v2acme.UnknownHostProblem,
			http.StatusBadRequest,
			"unknown host detail"},
		{
			v2acme.RateLimitedError("rate limited detail"),
			v2acme.RateLimitedProblem,
			http.StatusTooManyRequests,
			"rate limited detail"},
		{
			v2acme.BadNonceError("bad nonce detail"),
			v2acme.BadNonceProblem,
			http.StatusBadRequest,
			"bad nonce detail"},
		{
			v2acme.TLSError("TLS error detail"),
			v2acme.TLSProblem,
			http.StatusBadRequest,
			"TLS error detail"},
		{
			v2acme.RejectedIdentifierError("rejected identifier detail"),
			v2acme.RejectedIdentifierProblem,
			http.StatusBadRequest,
			"rejected identifier detail"},
		{
			v2acme.AccountDoesNotExistError("no account detail"),
			v2acme.AccountDoesNotExistProblem,
			http.StatusBadRequest,
			"no account detail"},
	}

	for _, c := range testCases {
		t.Run(string(c.pb.Type), func(t *testing.T) {
			assert.Equal(t, c.expectedType, c.pb.Type)
			assert.Equal(t, c.statusCode, c.pb.HTTPStatus)
			assert.Equal(t, c.detail, c.pb.Detail)
		})
	}
}

func Test_FromHTTPError(t *testing.T) {
	errConnRefused := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{
			Syscall: "getsockopt",
			Err:     syscall.ECONNREFUSED,
		},
	}

	errConnReset := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{
			Syscall: "getsockopt",
			Err:     syscall.ECONNRESET,
		},
	}

	cases := []struct {
		err      error
		expected string
	}{
		{
			errConnRefused,
			"connection refused",
		},
		{
			errConnReset,
			"connection reset by peer",
		},
		{
			errors.Trace(errConnRefused),
			"connection refused",
		},
		{
			errors.Trace(errConnReset),
			"connection reset by peer",
		},
		{
			errors.Annotate(errConnRefused, "some annotation"),
			"connection refused",
		},
		{
			errors.Annotate(errConnReset, "some annotation"),
			"connection reset by peer",
		},
	}
	for _, tc := range cases {
		prob := v2acme.ProblemFromHTTPError(tc.err)
		assert.Equal(t, tc.expected, prob.Detail)
	}
}
