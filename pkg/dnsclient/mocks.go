package dnsclient

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// Mock is a mock
type Mock struct {
	// KeyAuthorizationFile to be returned for succesful responses
	// base64(sha256("LoqXcYV8q5ONbJQxbmR7SCTNo3tiAXDfowyjxAjEuX0"
	//               + "." + "9jg46WB3rR_AHD-EBXdN7cBkH1WOu0tA3M9fm21mqTI"))
	// expected token + test account jwk thumbprint
	KeyAuthorizationFile string
}

// LookupTXT is a mock
func (mock *Mock) LookupTXT(_ context.Context, hostname string) ([]string, []string, error) {
	if hostname == "_acme-challenge.servfail.com" {
		return nil, nil, errors.New("SERVFAIL")
	}
	if hostname == "_acme-challenge.good-dns01.com" {
		return []string{mock.KeyAuthorizationFile}, []string{"respect my authority!"}, nil
	}
	if hostname == "_acme-challenge.wrong-dns01.com" {
		return []string{"a"}, []string{"respect my authority!"}, nil
	}
	if hostname == "_acme-challenge.wrong-many-dns01.com" {
		return []string{"a", "b", "c", "d", "e"}, []string{"respect my authority!"}, nil
	}
	if hostname == "_acme-challenge.long-dns01.com" {
		return []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, []string{"respect my authority!"}, nil
	}
	if hostname == "_acme-challenge.no-authority-dns01.com" {
		// base64(sha256("LoqXcYV8q5ONbJQxbmR7SCTNo3tiAXDfowyjxAjEuX0"
		//               + "." + "9jg46WB3rR_AHD-EBXdN7cBkH1WOu0tA3M9fm21mqTI"))
		// expected token + test account jwk thumbprint
		return []string{mock.KeyAuthorizationFile}, nil, nil
	}
	// empty-txts.com always returns zero TXT records
	if hostname == "_acme-challenge.empty-txts.com" {
		return []string{}, nil, nil
	}
	return []string{"hostname"}, []string{"respect my authority!"}, nil
}

// MockTimeoutError returns a a net.OpError for which Timeout() returns true.
func MockTimeoutError() *net.OpError {
	return &net.OpError{
		Err: os.NewSyscallError("ugh timeout", timeoutError{}),
	}
}

type timeoutError struct{}

func (t timeoutError) Error() string {
	return "so sloooow"
}
func (t timeoutError) Timeout() bool {
	return true
}

// LookupHost is a mock
//
// Note: see comments on LookupMX regarding email.only
func (mock *Mock) LookupHost(_ context.Context, hostname string) ([]net.IP, error) {
	if hostname == "always.invalid" ||
		hostname == "invalid.invalid" ||
		hostname == "email.only" {
		return []net.IP{}, nil
	}
	if hostname == "always.timeout" {
		return []net.IP{}, &Error{dns.TypeA, "always.timeout", MockTimeoutError(), -1}
	}
	if hostname == "always.error" {
		return []net.IP{}, &Error{dns.TypeA, "always.error", &net.OpError{
			Err: errors.New("some net error"),
		}, -1}
	}
	// dual-homed host with an IPv6 and an IPv4 address
	if hostname == "ipv4.and.ipv6.localhost" {
		return []net.IP{
			net.ParseIP("::1"),
			net.ParseIP("127.0.0.1"),
		}, nil
	}
	if hostname == "ipv6.localhost" {
		return []net.IP{
			net.ParseIP("::1"),
		}, nil
	}
	ip := net.ParseIP("127.0.0.1")
	return []net.IP{ip}, nil
}

// LookupCAA returns mock records for use in tests.
func (mock *Mock) LookupCAA(_ context.Context, domain string) ([]*dns.CAA, error) {
	return nil, nil
}

// LookupMX is a mock
//
// Note: the email.only domain must have an MX but no A or AAAA
// records. The mock LookupHost returns an address of 127.0.0.1 for
// all domains except for special cases, so MX-only domains must be
// handled in both LookupHost and LookupMX.
func (mock *Mock) LookupMX(_ context.Context, domain string) ([]string, error) {
	switch strings.TrimRight(domain, ".") {
	case "digicert.com", "certcentral.com":
		fallthrough
	case "email.only":
		fallthrough
	case "email.com":
		return []string{"mail.email.com"}, nil
	case "always.error":
		return []string{}, &Error{dns.TypeA, "always.error",
			&net.OpError{Err: errors.New("always.error always errors")}, -1}
	case "always.timeout":
		return []string{}, &Error{dns.TypeA, "always.timeout", MockTimeoutError(), -1}
	}
	return nil, nil
}
