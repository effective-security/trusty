package transport

import (
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTLSListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tlsInfo := &TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: trustedCAFile,
	}
	defer tlsInfo.Close()

	tlsln, err := NewTLSListener(ln, tlsInfo)
	require.NoError(t, err)

	t.Logf("listening on %v", tlsln.Addr().String())

	router := rest.NewRouter(notFoundHandler)

	srv := &http.Server{
		Handler:   router.Handler(),
		TLSConfig: tlsInfo.Config(),
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.Serve(tlsln)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		t.Logf("sending to %v", tlsln.Addr().String())
		res, err := http.Get("https://" + tlsln.Addr().String())
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "certificate signed by unknown authority")
			t.Logf("error from %v: %s", tlsln.Addr().String(), err.Error())
		}
		if res != nil {
			t.Logf("response code from %v: %d", tlsln.Addr().String(), res.StatusCode)
		}
	}()

	time.Sleep(3 * time.Second)
	tlsln.Close()
	wg.Wait()
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	marshal.WriteJSON(w, r, httperror.WithNotFound(r.URL.Path))
}
