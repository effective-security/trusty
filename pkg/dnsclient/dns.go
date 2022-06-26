package dnsclient

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/effective-security/metrics"
	"github.com/effective-security/xlog"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/pkg", "dnsclient")

// Resolver queries for DNS records
type Resolver interface {
	LookupTXT(context.Context, string) (txts []string, authorities []string, err error)
	LookupHost(context.Context, string) ([]net.IP, error)
	LookupCAA(context.Context, string) ([]*dns.CAA, error)
	LookupMX(context.Context, string) ([]string, error)
}

type exchanger interface {
	Exchange(m *dns.Msg, a string) (*dns.Msg, time.Duration, error)
}

// Client represents a client that talks to an external resolver
type Client struct {
	dnsClient                exchanger
	servers                  []string
	allowRestrictedAddresses bool
	maxTries                 int
}

var _ Resolver = &Client{}

// New constructs a new DNS resolver object that utilizes the
// provided list of DNS servers for resolution.
func New(servers []string, readTimeout time.Duration, maxTries int) *Client {
	// TODO(jmhodges): make constructor use an Option func pattern
	dnsClient := new(dns.Client)

	// Set timeout for underlying net.Conn
	dnsClient.ReadTimeout = readTimeout
	dnsClient.Net = "udp"

	return &Client{
		dnsClient:                dnsClient,
		servers:                  servers,
		allowRestrictedAddresses: false,
		maxTries:                 maxTries,
	}
}

// WithRestrictedAddresses will allow loopback addresses for TESTING purposes.
// This method should *ONLY* be called from tests (unit or integration).
func (c *Client) WithRestrictedAddresses() *Client {
	logger.Notice("WARNING=use_for_test_only")
	c.allowRestrictedAddresses = true
	return c
}

var (
	metricsKeyForTotalLookup    = []string{"dnsclient", "perf", "lookup"}
	metricsKeyForQuery          = []string{"dnsclient", "perf", "query"}
	metricsKeyForTimeoutCounter = []string{"dnsclient", "timeout", "query"}
)

// exchangeOne performs a single DNS exchange with a randomly chosen server
// out of the server list, returning the response, time, and error (if any).
// We assume that the upstream resolver requests and validates DNSSEC records
// itself.
func (c *Client) exchangeOne(ctx context.Context, hostname string, qtype uint16) (resp *dns.Msg, err error) {
	m := new(dns.Msg)
	// Set question type
	m.SetQuestion(dns.Fqdn(hostname), qtype)
	// Set the AD bit in the query header so that the resolver knows that
	// we are interested in this bit in the response header. If this isn't
	// set the AD bit in the response is useless (RFC 6840 Section 5.7).
	// This has no security implications, it simply allows us to gather
	// metrics about the percentage of responses that are secured with
	// DNSSEC.
	m.AuthenticatedData = true
	// Tell the resolver that we're willing to receive responses up to 4096 bytes.
	// This happens sometimes when there are a very large number of CAA records
	// present.
	m.SetEdns0(4096, false)

	if len(c.servers) < 1 {
		return nil, errors.Errorf("not configured with at least one DNS Server")
	}

	// Randomly pick a server
	chosenServerIndex := rand.Intn(len(c.servers))
	chosenServer := c.servers[chosenServerIndex]

	tries := 1
	qtypeStr := dns.TypeToString[qtype]
	started := time.Now().UTC()
	// Publish metrics
	defer func() {
		result := "failed"
		if resp != nil {
			result = dns.RcodeToString[resp.Rcode]
		}

		metrics.MeasureSince(
			metricsKeyForTotalLookup,
			started,
			metrics.Tag{Name: "qtype", Value: qtypeStr},
			metrics.Tag{Name: "result", Value: result},
			metrics.Tag{Name: "retries", Value: strconv.Itoa(tries)},
			metrics.Tag{Name: "resolver", Value: chosenServer},
		)
	}()

	type dnsResp struct {
		m   *dns.Msg
		err error
	}

	for {
		ch := make(chan dnsResp, 1)

		go func() {
			logger.Tracef("hostname=%q, type=%v, server=%s", hostname, qtypeStr, chosenServer)

			started := time.Now().UTC()
			rsp, _, err := c.dnsClient.Exchange(m, chosenServer)
			result := "failed"
			if rsp != nil {
				result = dns.RcodeToString[rsp.Rcode]
			}
			metrics.MeasureSince(
				metricsKeyForQuery,
				started,
				metrics.Tag{Name: "qtype", Value: qtypeStr},
				metrics.Tag{Name: "result", Value: result},
				metrics.Tag{Name: "resolver", Value: chosenServer},
			)

			ch <- dnsResp{m: rsp, err: err}
		}()
		select {
		case <-ctx.Done():
			reason := "unknown"
			err = ctx.Err()
			if err == context.DeadlineExceeded {
				reason = "deadline"
			} else if err == context.Canceled {
				reason = "canceled"
			}
			metrics.IncrCounter(
				metricsKeyForTimeoutCounter,
				1,
				metrics.Tag{Name: "qtype", Value: qtypeStr},
				metrics.Tag{Name: "reason", Value: reason},
				metrics.Tag{Name: "resolver", Value: chosenServer},
			)

			if err != nil {
				logger.Errorf("hostname=%q, type=%v, err=[%+v]", hostname, qtypeStr, err.Error())
			} else {
				logger.Warningf("hostname=%q, type=%v", hostname, qtypeStr)
			}
			return
		case r := <-ch:
			if r.err != nil {
				operr, ok := r.err.(*net.OpError)
				isRetryable := ok && operr.Temporary()
				hasRetriesLeft := tries < c.maxTries
				if isRetryable && hasRetriesLeft {
					tries++
					// Chose a new server to retry the query with by incrementing the
					// chosen server index modulo the number of servers. This ensures that
					// if one dns server isn't available we retry with the next in the
					// list.
					chosenServerIndex = (chosenServerIndex + 1) % len(c.servers)
					chosenServer = c.servers[chosenServerIndex]
					continue
				} else if isRetryable && !hasRetriesLeft {
					logger.Errorf("reason=out_of_retries, hostname=%q, type=%v", hostname, qtypeStr)
				}
			}
			resp, err = r.m, r.err
			return
		}
	}
}

// LookupTXT sends a DNS query to find all TXT records associated with
// the provided hostname which it returns along with the returned
// DNS authority section.
func (c *Client) LookupTXT(ctx context.Context, hostname string) ([]string, []string, error) {
	var txt []string
	dnsType := dns.TypeTXT
	r, err := c.exchangeOne(ctx, hostname, dnsType)
	if err != nil {
		logger.Tracef("reason=exchangeOne, host=%s, err=[%+v]", hostname, err.Error())
		return nil, nil, &Error{dnsType, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		logger.Tracef("reason=exchangeOne, host=%s, rc=%d", hostname, r.Rcode)
		return nil, nil, &Error{dnsType, hostname, nil, r.Rcode}
	}

	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dnsType {
			if txtRec, ok := answer.(*dns.TXT); ok {
				txt = append(txt, strings.Join(txtRec.Txt, ""))
			}
		}
	}

	authorities := []string{}
	for _, a := range r.Ns {
		authorities = append(authorities, a.String())
	}

	return txt, authorities, err
}

func (c *Client) lookupIP(ctx context.Context, hostname string, ipType uint16) ([]dns.RR, error) {
	resp, err := c.exchangeOne(ctx, hostname, ipType)
	if err != nil {
		logger.Tracef("reason=exchangeOne, type=%d, host=%s, err=[%+v]", ipType, hostname, err.Error())
		return nil, &Error{ipType, hostname, err, -1}
	}
	if resp.Rcode != dns.RcodeSuccess {
		logger.Tracef("reason=exchangeOne, type=%d, host=%s, rc=%d", ipType, hostname, resp.Rcode)
		return nil, &Error{ipType, hostname, nil, resp.Rcode}
	}
	return resp.Answer, nil
}

// LookupHost sends a DNS query to find all A and AAAA records associated with
// the provided hostname. This method assumes that the external resolver will
// chase CNAME/DNAME aliases and return relevant records.  It will retry
// requests in the case of temporary network errors. It can return net package,
// context.Canceled, and context.DeadlineExceeded errors, all wrapped in the
// Error type.
func (c *Client) LookupHost(ctx context.Context, hostname string) ([]net.IP, error) {
	var recordsA, recordsAAAA []dns.RR
	var errA, errAAAA error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		recordsA, errA = c.lookupIP(ctx, hostname, dns.TypeA)
		if errA != nil {
			logger.Tracef("reason=lookupIP, type=A, host=%s, err=[%+v]",
				hostname, errA.Error())
		} else {
			logger.Tracef("reason=lookupIP, type=A, host=%s, records=%d",
				hostname, len(recordsA))
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		recordsAAAA, errAAAA = c.lookupIP(ctx, hostname, dns.TypeAAAA)
		if errAAAA != nil {
			logger.Tracef("reason=lookupIP, type=AAAA, host=%s, err=[%+v]", hostname, errAAAA.Error())
		} else {
			logger.Tracef("reason=lookupIP, type=AAAA, host=%s, records=%d",
				hostname, len(recordsAAAA))
		}
	}()
	wg.Wait()

	if errA != nil && errAAAA != nil {
		return nil, errors.WithStack(errA)
	}

	var addrs []net.IP

	for _, answer := range recordsA {
		if answer.Header().Rrtype == dns.TypeA {
			if a, ok := answer.(*dns.A); ok &&
				a.A.To4() != nil &&
				(!isPrivateV4(a.A) || c.allowRestrictedAddresses) {
				addrs = append(addrs, a.A)
			}
		}
	}
	for _, answer := range recordsAAAA {
		if answer.Header().Rrtype == dns.TypeAAAA {
			if aaaa, ok := answer.(*dns.AAAA); ok &&
				aaaa.AAAA.To16() != nil &&
				(!isPrivateV6(aaaa.AAAA) || c.allowRestrictedAddresses) {
				addrs = append(addrs, aaaa.AAAA)
			}
		}
	}

	return addrs, nil
}

// LookupCAA sends a DNS query to find all CAA records associated with
// the provided hostname.
func (c *Client) LookupCAA(ctx context.Context, hostname string) ([]*dns.CAA, error) {
	dnsType := dns.TypeCAA
	r, err := c.exchangeOne(ctx, hostname, dnsType)
	if err != nil {
		logger.Tracef("reason=exchangeOne, host=%s, err=[%+v]", hostname, err.Error())
		return nil, &Error{dnsType, hostname, err, -1}
	}

	if r.Rcode == dns.RcodeServerFailure {
		logger.Tracef("reason=exchangeOne, host=%s, rc=%d", hostname, r.Rcode)
		return nil, &Error{dnsType, hostname, nil, r.Rcode}
	}

	var CAAs []*dns.CAA
	for _, answer := range r.Answer {
		if caaR, ok := answer.(*dns.CAA); ok {
			CAAs = append(CAAs, caaR)
		}
	}
	return CAAs, nil
}

// LookupMX sends a DNS query to find a MX record associated hostname and returns the
// record target.
func (c *Client) LookupMX(ctx context.Context, hostname string) ([]string, error) {
	dnsType := dns.TypeMX
	r, err := c.exchangeOne(ctx, hostname, dnsType)
	if err != nil {
		logger.Tracef("reason=exchangeOne, host=%s, err=[%+v]", hostname, err.Error())
		return nil, &Error{dnsType, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		logger.Tracef("reason=exchangeOne, host=%s, rc=%d", hostname, r.Rcode)
		return nil, &Error{dnsType, hostname, nil, r.Rcode}
	}

	var results []string
	for _, answer := range r.Answer {
		if mx, ok := answer.(*dns.MX); ok {
			results = append(results, mx.Mx)
		}
	}

	return results, nil
}
