package dnsclient

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/effective-security/xlog"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dnsLoopbackAddr = "127.0.0.1:4053"

func mockDNSQuery(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	appendAnswer := func(rr dns.RR) {
		m.Answer = append(m.Answer, rr)
	}
	for _, q := range r.Question {
		q.Name = strings.ToLower(q.Name)
		if q.Name == "servfail.com." || q.Name == "servfailexception.example.com" {
			m.Rcode = dns.RcodeServerFailure
			break
		}
		switch q.Qtype {
		case dns.TypeSOA:
			record := new(dns.SOA)
			record.Hdr = dns.RR_Header{Name: "certcentral.com.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 0}
			record.Ns = "ns.certcentral.com."
			record.Mbox = "master.certcentral.com."
			record.Serial = 1
			record.Refresh = 1
			record.Retry = 1
			record.Expire = 1
			record.Minttl = 1
			appendAnswer(record)
		case dns.TypeAAAA:
			if q.Name == "v6.certcentral.com." {
				record := new(dns.AAAA)
				record.Hdr = dns.RR_Header{Name: "v6.certcentral.com.", Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
				record.AAAA = net.ParseIP("::1")
				appendAnswer(record)
			}
			if q.Name == "dualstack.certcentral.com." {
				record := new(dns.AAAA)
				record.Hdr = dns.RR_Header{Name: "dualstack.certcentral.com.", Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
				record.AAAA = net.ParseIP("::1")
				appendAnswer(record)
			}
			if q.Name == "v4error.certcentral.com." {
				record := new(dns.AAAA)
				record.Hdr = dns.RR_Header{Name: "v4error.certcentral.com.", Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
				record.AAAA = net.ParseIP("::1")
				appendAnswer(record)
			}
			if q.Name == "v6error.certcentral.com." {
				m.SetRcode(r, dns.RcodeNotImplemented)
			}
			if q.Name == "nxdomain.certcentral.com." {
				m.SetRcode(r, dns.RcodeNameError)
			}
			if q.Name == "dualstackerror.certcentral.com." {
				m.SetRcode(r, dns.RcodeNotImplemented)
			}
		case dns.TypeA:
			if q.Name == "cps.certcentral.com." {
				record := new(dns.A)
				record.Hdr = dns.RR_Header{Name: "cps.certcentral.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
				record.A = net.ParseIP("127.0.0.1")
				appendAnswer(record)
			}
			if q.Name == "dualstack.certcentral.com." {
				record := new(dns.A)
				record.Hdr = dns.RR_Header{Name: "dualstack.certcentral.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
				record.A = net.ParseIP("127.0.0.1")
				appendAnswer(record)
			}
			if q.Name == "v6error.certcentral.com." {
				record := new(dns.A)
				record.Hdr = dns.RR_Header{Name: "dualstack.certcentral.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
				record.A = net.ParseIP("127.0.0.1")
				appendAnswer(record)
			}
			if q.Name == "v4error.certcentral.com." {
				m.SetRcode(r, dns.RcodeNotImplemented)
			}
			if q.Name == "nxdomain.certcentral.com." {
				m.SetRcode(r, dns.RcodeNameError)
			}
			if q.Name == "dualstackerror.certcentral.com." {
				m.SetRcode(r, dns.RcodeRefused)
			}
		case dns.TypeCNAME:
			if q.Name == "cname.certcentral.com." {
				record := new(dns.CNAME)
				record.Hdr = dns.RR_Header{Name: "cname.certcentral.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 30}
				record.Target = "cps.certcentral.com."
				appendAnswer(record)
			}
			if q.Name == "cname.example.com." {
				record := new(dns.CNAME)
				record.Hdr = dns.RR_Header{Name: "cname.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 30}
				record.Target = "CAA.example.com."
				appendAnswer(record)
			}
		case dns.TypeDNAME:
			if q.Name == "dname.certcentral.com." {
				record := new(dns.DNAME)
				record.Hdr = dns.RR_Header{Name: "dname.certcentral.com.", Rrtype: dns.TypeDNAME, Class: dns.ClassINET, Ttl: 30}
				record.Target = "cps.certcentral.com."
				appendAnswer(record)
			}
		case dns.TypeCAA:
			if q.Name == "bracewel.net." || q.Name == "caa.example.com." {
				record := new(dns.CAA)
				record.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeCAA, Class: dns.ClassINET, Ttl: 0}
				record.Tag = "issue"
				record.Value = "certcentral.com"
				record.Flag = 1
				appendAnswer(record)
			}
			if q.Name == "cname.example.com." {
				record := new(dns.CAA)
				record.Hdr = dns.RR_Header{Name: "caa.example.com.", Rrtype: dns.TypeCAA, Class: dns.ClassINET, Ttl: 0}
				record.Tag = "issue"
				record.Value = "certcentral.com"
				record.Flag = 1
				appendAnswer(record)
			}
		case dns.TypeTXT:
			if q.Name == "split-txt.certcentral.com." {
				record := new(dns.TXT)
				record.Hdr = dns.RR_Header{Name: "split-txt.certcentral.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
				record.Txt = []string{"a", "b", "c"}
				appendAnswer(record)
			} else {
				auth := new(dns.SOA)
				auth.Hdr = dns.RR_Header{Name: "certcentral.com.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 0}
				auth.Ns = "ns.certcentral.com."
				auth.Mbox = "master.certcentral.com."
				auth.Serial = 1
				auth.Refresh = 1
				auth.Retry = 1
				auth.Expire = 1
				auth.Minttl = 1
				m.Ns = append(m.Ns, auth)
			}
			if q.Name == "nxdomain.certcentral.com." {
				m.SetRcode(r, dns.RcodeNameError)
			}
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		panic(err) // running tests, so panic is OK
	}
	return
}

func serveLoopResolver(stopChan chan bool) {
	dns.HandleFunc(".", mockDNSQuery)
	tcpServer := &dns.Server{
		Addr:         dnsLoopbackAddr,
		Net:          "tcp",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
	udpServer := &dns.Server{
		Addr:         dnsLoopbackAddr,
		Net:          "udp",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
	go func() {
		err := tcpServer.ListenAndServe()
		if err != nil {
			logger.KV(xlog.ERROR, "reason", "tcpServer.ListenAndServe", "err", err)
		}
	}()
	go func() {
		err := udpServer.ListenAndServe()
		if err != nil {
			logger.KV(xlog.ERROR, "reason", "udpServer.ListenAndServe", "err", err)
		}
	}()
	go func() {
		<-stopChan
		err := tcpServer.Shutdown()
		if err != nil {
			logger.Fatalf("reason=tcpServer.Shutdown, err=[%+v]", err)
		}
		err = udpServer.Shutdown()
		if err != nil {
			logger.Fatalf("reason=udpServer.Shutdown, err=[%+v]", err)
		}
	}()
}

func pollServer() {
	backoff := time.Duration(200 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	ticker := time.NewTicker(backoff)

	for {
		select {
		case <-ctx.Done():
			logger.Panicf("reason='timeout reached while testing for the dns server to come up'")
		case <-ticker.C:
			conn, _ := dns.DialTimeout("udp", dnsLoopbackAddr, backoff)
			if conn != nil {
				_ = conn.Close()
				return
			}
		}
	}
}

func TestMain(m *testing.M) {
	stop := make(chan bool, 1)
	serveLoopResolver(stop)
	pollServer()
	ret := m.Run()
	stop <- true
	os.Exit(ret)
}

func Test_DNSNoServers(t *testing.T) {
	obj := New([]string{}, time.Hour, 1).WithRestrictedAddresses()

	_, err := obj.LookupHost(context.Background(), "certcentral.com")
	assert.Error(t, err)
}

func Test_DNSOneServer(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	_, err := obj.LookupHost(context.Background(), "certcentral.com")
	assert.NoError(t, err)
}

func Test_DNSDuplicateServers(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr, dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	_, err := obj.LookupHost(context.Background(), "certcentral.com")
	assert.NoError(t, err)
}

func Test_DNSLookupsNoServer(t *testing.T) {
	obj := New([]string{}, time.Second*10, 1).WithRestrictedAddresses()

	_, _, err := obj.LookupTXT(context.Background(), "certcentral.com")
	assert.Error(t, err)

	_, err = obj.LookupHost(context.Background(), "certcentral.com")
	assert.Error(t, err)

	_, err = obj.LookupCAA(context.Background(), "certcentral.com")
	assert.Error(t, err)
}

func Test_DNSServFail(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()
	bad := "servfail.com"

	_, _, err := obj.LookupTXT(context.Background(), bad)
	assert.Error(t, err, "LookupTXT didn't return an error")

	_, err = obj.LookupHost(context.Background(), bad)
	assert.Error(t, err, "LookupHost didn't return an error")

	emptyCaa, err := obj.LookupCAA(context.Background(), bad)
	assert.Empty(t, emptyCaa, "Query returned non-empty list of CAA records")
	assert.Error(t, err, "LookupCAA should have returned an error")
}

func Test_DNSLookupTXT(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	a, _, err := obj.LookupTXT(context.Background(), "certcentral.com")
	t.Logf("A: %v", a)
	assert.NoError(t, err, "No message")

	a, _, err = obj.LookupTXT(context.Background(), "split-txt.certcentral.com")
	t.Logf("A: %v ", a)
	assert.NoError(t, err, "No message")
	require.Equal(t, 1, len(a))
	assert.Equal(t, "abc", a[0])
}

func Test_DNSLookupHost(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	ip, err := obj.LookupHost(context.Background(), "servfail.com")
	t.Logf("servfail.com - IP: %s, Err: %v", ip, err)
	assert.Error(t, err, "Server failure")
	assert.Empty(t, ip, "Should not have IPs")

	ip, err = obj.LookupHost(context.Background(), "nonexistent.certcentral.com")
	t.Logf("nonexistent.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to not exist")
	assert.Empty(t, ip, "Should not have IPs")

	// Single IPv4 address
	ip, err = obj.LookupHost(context.Background(), "cps.certcentral.com")
	t.Logf("cps.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	assert.Equal(t, 1, len(ip), "Should have IP")
	ip, err = obj.LookupHost(context.Background(), "cps.certcentral.com")
	t.Logf("cps.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	assert.Equal(t, 1, len(ip), "Should have IP")

	// Single IPv6 address
	ip, err = obj.LookupHost(context.Background(), "v6.certcentral.com")
	t.Logf("v6.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	assert.Equal(t, 1, len(ip), "Should not have IPs")

	// Both IPv6 and IPv4 address
	ip, err = obj.LookupHost(context.Background(), "dualstack.certcentral.com")
	t.Logf("dualstack.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	require.Equal(t, 2, len(ip), "Should have 2 IPs")
	expected := net.ParseIP("127.0.0.1")
	assert.True(t, ip[0].To4().Equal(expected), "wrong ipv4 address")
	expected = net.ParseIP("::1")
	assert.True(t, ip[1].To16().Equal(expected), "wrong ipv6 address")

	// IPv6 error, IPv4 success
	ip, err = obj.LookupHost(context.Background(), "v6error.certcentral.com")
	t.Logf("v6error.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	require.Equal(t, 1, len(ip), "Should have 1 IP")
	expected = net.ParseIP("127.0.0.1")
	assert.True(t, ip[0].To4().Equal(expected), "wrong ipv4 address")

	// IPv6 success, IPv4 error
	ip, err = obj.LookupHost(context.Background(), "v4error.certcentral.com")
	t.Logf("v4error.certcentral.com - IP: %s, Err: %v", ip, err)
	assert.NoError(t, err, "Not an error to exist")
	require.Equal(t, 1, len(ip), "Should have 1 IP")
	expected = net.ParseIP("::1")
	assert.True(t, ip[0].To16().Equal(expected), "wrong ipv6 address")

	// IPv6 error, IPv4 error
	// Should return the IPv4 error (Refused) and not IPv6 error (NotImplemented)
	hostname := "dualstackerror.certcentral.com"
	ip, err = obj.LookupHost(context.Background(), hostname)
	t.Logf("%s - IP: %s, Err: %v", hostname, ip, err)
	assert.Error(t, err, "Should be an error")
	expectedErr := Error{dns.TypeA, hostname, nil, dns.RcodeRefused}
	if err, ok := errors.Cause(err).(*Error); !ok || *err != expectedErr {
		t.Errorf("Looking up %s, got %#v, expected %#v", hostname, err, expectedErr)
	}
}

func Test_DNSNXDOMAIN(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	hostname := "nxdomain.certcentral.com"
	_, err := obj.LookupHost(context.Background(), hostname)
	expected := Error{dns.TypeA, hostname, nil, dns.RcodeNameError}
	if err, ok := errors.Cause(err).(*Error); !ok || *err != expected {
		t.Errorf("Looking up %s, got %#v, expected %#v", hostname, err, expected)
	}

	_, _, err = obj.LookupTXT(context.Background(), hostname)
	expected.recordType = dns.TypeTXT
	if err, ok := errors.Cause(err).(*Error); !ok || *err != expected {
		t.Errorf("Looking up %s, got %#v, expected %#v", hostname, err, expected)
	}
}

func Test_DNSLookupCAA(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	caas, err := obj.LookupCAA(context.Background(), "bracewel.net")
	assert.NoError(t, err, "CAA lookup failed")
	assert.True(t, len(caas) > 0, "Should have CAA records")

	caas, err = obj.LookupCAA(context.Background(), "nonexistent.certcentral.com")
	assert.NoError(t, err, "CAA lookup failed")
	assert.True(t, len(caas) == 0, "Shouldn't have CAA records")

	caas, err = obj.LookupCAA(context.Background(), "cname.example.com")
	assert.NoError(t, err, "CAA lookup failed")
	assert.True(t, len(caas) > 0, "Should follow CNAME to find CAA")
}

func TestDNSTXTAuthorities(t *testing.T) {
	obj := New([]string{dnsLoopbackAddr}, time.Second*10, 1).WithRestrictedAddresses()

	_, auths, err := obj.LookupTXT(context.Background(), "certcentral.com")

	assert.NoError(t, err, "TXT lookup failed")
	assert.Equal(t, 1, len(auths))
	assert.Equal(t, "certcentral.com.	0	IN	SOA	ns.certcentral.com. master.certcentral.com. 1 1 1 1 1", auths[0])
}

func TestIsPrivateIP(t *testing.T) {
	assert.True(t, isPrivateV4(net.ParseIP("127.0.0.1")), "should be private")
	assert.True(t, isPrivateV4(net.ParseIP("192.168.254.254")), "should be private")
	assert.True(t, isPrivateV4(net.ParseIP("10.255.0.3")), "should be private")
	assert.True(t, isPrivateV4(net.ParseIP("172.16.255.255")), "should be private")
	assert.True(t, isPrivateV4(net.ParseIP("172.31.255.255")), "should be private")
	assert.True(t, !isPrivateV4(net.ParseIP("128.0.0.1")), "should be private")
	assert.True(t, !isPrivateV4(net.ParseIP("192.169.255.255")), "should not be private")
	assert.True(t, !isPrivateV4(net.ParseIP("9.255.0.255")), "should not be private")
	assert.True(t, !isPrivateV4(net.ParseIP("172.32.255.255")), "should not be private")

	assert.True(t, isPrivateV6(net.ParseIP("::0")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("::1")), "should be private")
	assert.True(t, !isPrivateV6(net.ParseIP("::2")), "should not be private")

	assert.True(t, isPrivateV6(net.ParseIP("fe80::1")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("febf::1")), "should be private")
	assert.False(t, isPrivateV6(net.ParseIP("fec0::1")), "should not be private")
	assert.False(t, isPrivateV6(net.ParseIP("feff::1")), "should not be private")

	assert.True(t, isPrivateV6(net.ParseIP("ff00::1")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("ff10::1")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")), "should be private")

	assert.True(t, isPrivateV6(net.ParseIP("2002::")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("2002:ffff:ffff:ffff:ffff:ffff:ffff:ffff")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("0100::")), "should be private")
	assert.True(t, isPrivateV6(net.ParseIP("0100::0000:ffff:ffff:ffff:ffff")), "should be private")
	assert.False(t, isPrivateV6(net.ParseIP("0100::0001:0000:0000:0000:0000")), "should be private")
}

type testExchanger struct {
	sync.Mutex
	count int
	errs  []error
}

var errTooManyRequests = errors.New("too many requests")

func (te *testExchanger) Exchange(m *dns.Msg, a string) (*dns.Msg, time.Duration, error) {
	te.Lock()
	defer te.Unlock()
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{Rcode: dns.RcodeSuccess},
	}
	if len(te.errs) <= te.count {
		return nil, 0, errTooManyRequests
	}
	err := te.errs[te.count]
	te.count++

	return msg, 2 * time.Millisecond, err
}

func Test_Retry(t *testing.T) {
	isTempErr := &net.OpError{Op: "read", Err: tempError(true)}
	nonTempErr := &net.OpError{Op: "read", Err: tempError(false)}
	servFailError := errors.New("DNS problem: server failure at resolver looking up TXT for example.com")
	netError := errors.New("DNS problem: networking error looking up TXT for example.com")
	type testCase struct {
		maxTries      int
		te            *testExchanger
		expected      error
		expectedCount int
	}
	tests := []*testCase{
		// The success on first try case
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{nil},
			},
			expected:      nil,
			expectedCount: 1,
		},
		// Immediate non-OpError, error returns immediately
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{errors.New("nope")},
			},
			expected:      servFailError,
			expectedCount: 1,
		},
		// Temporary err, then non-OpError stops at two tries
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{isTempErr, errors.New("nope")},
			},
			expected:      servFailError,
			expectedCount: 2,
		},
		// Temporary error given always
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{
					isTempErr,
					isTempErr,
					isTempErr,
				},
			},
			expected:      netError,
			expectedCount: 3,
		},
		// Even with maxTries at 0, we should still let a single request go
		// through
		{
			maxTries: 0,
			te: &testExchanger{
				errs: []error{nil},
			},
			expected:      nil,
			expectedCount: 1,
		},
		// Temporary error given just once causes two tries
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{
					isTempErr,
					nil,
				},
			},
			expected:      nil,
			expectedCount: 2,
		},
		// Temporary error given twice causes three tries
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{
					isTempErr,
					isTempErr,
					nil,
				},
			},
			expected:      nil,
			expectedCount: 3,
		},
		// Temporary error given thrice causes three tries and fails
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{
					isTempErr,
					isTempErr,
					isTempErr,
				},
			},
			expected:      netError,
			expectedCount: 3,
		},
		// temporary then non-Temporary error causes two retries
		{
			maxTries: 3,
			te: &testExchanger{
				errs: []error{
					isTempErr,
					nonTempErr,
				},
			},
			expected:      netError,
			expectedCount: 2,
		},
	}

	for i, tc := range tests {
		dr := New([]string{dnsLoopbackAddr}, time.Second*10, tc.maxTries).WithRestrictedAddresses()
		dr.dnsClient = tc.te
		_, _, err := dr.LookupTXT(context.Background(), "example.com")
		if err == errTooManyRequests {
			t.Errorf("#%d, sent more requests than the test case handles", i)
		}
		expectedErr := tc.expected
		if (expectedErr == nil && err != nil) ||
			(expectedErr != nil && err == nil) ||
			(expectedErr != nil && expectedErr.Error() != err.Error()) {
			t.Errorf("#%d, error, expected %v, got %v", i, expectedErr, err)
		}
		if tc.expectedCount != tc.te.count {
			t.Errorf("#%d, error, expectedCount %v, got %v", i, tc.expectedCount, tc.te.count)
		}
	}

	dr := New([]string{dnsLoopbackAddr}, time.Second*10, 3).WithRestrictedAddresses()
	dr.dnsClient = &testExchanger{errs: []error{isTempErr, isTempErr, nil}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := dr.LookupTXT(ctx, "example.com")
	if err == nil ||
		err.Error() != "DNS problem: query timed out looking up TXT for example.com" {
		t.Errorf("expected %s, got %s", context.Canceled, err)
	}

	dr.dnsClient = &testExchanger{errs: []error{isTempErr, isTempErr, nil}}
	ctx, cancel = context.WithTimeout(context.Background(), -10*time.Hour)
	defer cancel()
	_, _, err = dr.LookupTXT(ctx, "example.com")
	if err == nil ||
		err.Error() != "DNS problem: query timed out looking up TXT for example.com" {
		t.Errorf("expected %s, got %s", context.DeadlineExceeded, err)
	}

	dr.dnsClient = &testExchanger{errs: []error{isTempErr, isTempErr, nil}}
	ctx, deadlineCancel := context.WithTimeout(context.Background(), -10*time.Hour)
	deadlineCancel()
	_, _, err = dr.LookupTXT(ctx, "example.com")
	if err == nil ||
		err.Error() != "DNS problem: query timed out looking up TXT for example.com" {
		t.Errorf("expected %s, got %s", context.DeadlineExceeded, err)
	}
}

type tempError bool

func (t tempError) Temporary() bool { return bool(t) }
func (t tempError) Error() string   { return fmt.Sprintf("Temporary: %t", t) }

// rotateFailureExchanger is a dns.Exchange implementation that tracks a count
// of the number of calls to `Exchange` for a given address in the `lookups`
// map. For all addresses in the `brokenAddresses` map, a retryable error is
// returned from `Exchange`. This mock is used by `TestRotateServerOnErr`.
type rotateFailureExchanger struct {
	sync.Mutex
	lookups         map[string]int
	brokenAddresses map[string]bool
}

// Exchange for rotateFailureExchanger tracks the `a` argument in `lookups` and
// if present in `brokenAddresses`, returns a temporary error.
func (e *rotateFailureExchanger) Exchange(m *dns.Msg, a string) (*dns.Msg, time.Duration, error) {
	e.Lock()
	defer e.Unlock()

	// Track that exchange was called for the given server
	e.lookups[a]++

	// If its a broken server, return a retryable error
	if e.brokenAddresses[a] {
		isTempErr := &net.OpError{Op: "read", Err: tempError(true)}
		return nil, 2 * time.Millisecond, isTempErr
	}

	return m, 2 * time.Millisecond, nil
}

// TestRotateServerOnErr ensures that a retryable error returned from a DNS
// server will result in the retry being performed against the next server in
// the list.
func TestRotateServerOnErr(t *testing.T) {
	// Configure three DNS servers
	dnsServers := []string{
		"a", "b", "c",
	}
	// Set up a DNS client using these servers that will retry queries up to
	// a maximum of 5 times. It's important to choose a maxTries value >= the
	// number of dnsServers to ensure we always get around to trying the one
	// working server
	maxTries := 5
	client := New(dnsServers, time.Second*10, maxTries).WithRestrictedAddresses()

	// Configure a mock exchanger that will always return a retryable error for
	// the A and B servers. This will force the C server to do all the work once
	// retries reach it.
	mock := &rotateFailureExchanger{
		brokenAddresses: map[string]bool{"a": true, "b": true},
		lookups:         make(map[string]int),
	}
	client.dnsClient = mock

	// Perform a bunch of lookups. We choose the initial server randomly. Any time
	// A or B is chosen there should be an error and a retry using the next server
	// in the list. Since we configured maxTries to be larger than the number of
	// servers *all* queries should eventually succeed by being retried against
	// the C server.
	for i := 0; i < maxTries*2; i++ {
		_, _, err := client.LookupTXT(context.Background(), "example.com")
		// Any errors are unexpected - the C server should have responded without error.
		assert.NoError(t, err, "Expected no error from eventual retry with functional server")
	}

	// We expect that the A and B servers had a non-zero number of lookups attempted
	assert.True(t, mock.lookups["a"] > 0, "Expected A server to have non-zero lookup attempts")
	assert.True(t, mock.lookups["b"] > 0, "Expected B server to have non-zero lookup attempts")
	// We expect that the C server eventually served all of the lookups attempted
	assert.Equal(t, maxTries*2, mock.lookups["c"])
}
