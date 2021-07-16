package dnsclient

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const (
	detailDNSTimeout    = "query timed out"
	detailDNSNetFailure = "networking error"
	detailServerFailure = "server failure at resolver"
)

// Error wraps a DNS error with various relevant information
type Error struct {
	recordType uint16
	hostname   string
	// Exactly one of `rc` or `underlying` should be set.
	underlying error
	rc         int
}

func (e Error) Error() string {
	var detail string
	if e.underlying != nil {
		if netErr, ok := e.underlying.(*net.OpError); ok {
			if netErr.Timeout() {
				detail = detailDNSTimeout
			} else {
				detail = detailDNSNetFailure
			}
			// Note: we check d.underlying here even though `Timeout()` does this because the call to `netErr.Timeout()` above only
			// happens for `*net.OpError` underlying types!
		} else if e.underlying == context.Canceled || e.underlying == context.DeadlineExceeded {
			detail = detailDNSTimeout
		} else {
			detail = detailServerFailure
		}
	} else if e.rc != dns.RcodeSuccess {
		detail = dns.RcodeToString[e.rc]
	} else {
		detail = detailServerFailure
	}
	return fmt.Sprintf("DNS problem: %s looking up %s for %s",
		detail, dns.TypeToString[e.recordType], e.hostname)
}

// Timeout returns true if the underlying error was a timeout
func (e Error) Timeout() bool {
	if netErr, ok := e.underlying.(*net.OpError); ok {
		return netErr.Timeout()
	} else if e.underlying == context.Canceled || e.underlying == context.DeadlineExceeded {
		return true
	}
	return false
}
