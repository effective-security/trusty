package acme

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
)

// NonceHandler returns nonce
func (s *Service) NonceHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		w.Header().Set(header.Link, link(s.baseURL()+v2acme.PathForDirectoryBase, "index"))
		s.handleACMEHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Service) handleACMEHeaders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || strings.Contains(r.URL.Path, uriNewNonce) {

		nonce, _ := s.Nonce()

		logger.KV(xlog.DEBUG,
			"method", r.Method,
			"path", r.URL.Path,
			"nonce", nonce)

		w.Header().Set(header.ReplayNonce, nonce)
		w.Header().Add(header.CacheControl, "public, max-age=0, no-cache")
	}
}

// Nonce returns nonce
func (s *Service) Nonce() (string, error) {
	now := time.Now()
	nonce, err := s.cadb.CreateNonce(context.Background(), &model.Nonce{
		Nonce:     certutil.RandomString(16),
		CreatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	})
	if err != nil {
		return "", errors.Trace(err)
	}
	return nonce.Nonce, nil
}
