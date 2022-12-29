package cis

import (
	"encoding/pem"
	"net/http"

	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/porto/xhttp/header"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/porto/xhttp/marshal"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/xlog"
)

/*
const (
	ikidTag = "ikid"
	skidTag = "skid"
)
*/
// GetCRLHandler returns CRL
func (s *Service) GetCRLHandler() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, p restserver.Params) {
		ikid := p.ByName("issuer_id")
		/*
			if strings.HasSuffix(ikid, ".crl") {
				ikid = ikid[0 : len(ikid)-4]
			}
		*/
		logger.KV(xlog.TRACE, "ikid", ikid)

		ctx := r.Context()
		m, err := s.db.GetCrl(ctx, ikid)
		if err != nil {
			if xdb.IsNotFoundError(err) {
				// metrics for Not Found
				metricskey.AIADownloadFailCrl.IncrCounter(1)
				marshal.WriteJSON(w, r, httperror.NotFound("unable to locate CRL"))

			} else {
				marshal.WriteJSON(w, r, httperror.Unexpected("unable to locate CRL").WithCause(err))
			}
			return
		}

		block, _ := pem.Decode([]byte(m.Pem))

		metricskey.AIADownloadSuccessCrl.IncrCounter(1)

		wh := w.Header()
		wh.Set(header.ContentType, "application/pkix-crl")
		wh.Set("Cache-Control", "no-store")
		wh.Set("Pragma", "no-cache")
		_, _ = w.Write(block.Bytes)
	}
}

// GetCertHandler returns certificate
func (s *Service) GetCertHandler() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, p restserver.Params) {
		skid := p.ByName("subject_id")
		/*
			if strings.HasSuffix(skid, ".crt") {
				skid = skid[0 : len(skid)-4]
			}
		*/
		logger.KV(xlog.TRACE, "skid", skid)

		ctx := r.Context()
		m, err := s.db.GetCertificateBySKID(ctx, skid)
		if err != nil {
			if xdb.IsNotFoundError(err) {
				// metrics for Not Found
				metricskey.AIADownloadFailCert.IncrCounter(1)
				marshal.WriteJSON(w, r, httperror.NotFound("unable to locate certificate"))

			} else {
				marshal.WriteJSON(w, r, httperror.Unexpected("unable to locate certificate").WithCause(err))
			}
			return
		}

		block, _ := pem.Decode([]byte(m.Pem))
		metricskey.AIADownloadSuccessCert.IncrCounter(1)

		w.Header().Set(header.ContentType, "application/pkix-cert")
		_, _ = w.Write(block.Bytes)
	}
}
