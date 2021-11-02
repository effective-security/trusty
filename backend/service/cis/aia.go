package cis

import (
	"encoding/pem"
	"net/http"

	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/martinisecurity/trusty/backend/db"
)

var (
	mkAIADownloadCertSuccessful = []string{"aia", "download", "cert", "successful"}
	mkAIADownloadCertFailed     = []string{"aia", "download", "cert", "failed"}
	mkAIADownloadCrlSuccessful  = []string{"aia", "download", "crl", "successful"}
	mkAIADownloadCrlFailed      = []string{"aia", "download", "crl", "failed"}
)

const (
	ikidTag = "ikid"
	skidTag = "skid"
)

// GetCRLHandler returns CRL
func (s *Service) GetCRLHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
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
			if db.IsNotFoundError(err) {
				// metrics for Not Found
				metrics.IncrCounter(mkAIADownloadCrlFailed, 1,
					metrics.Tag{Name: ikidTag, Value: ikid},
				)
				marshal.WriteJSON(w, r, httperror.WithNotFound("unable to locate CRL"))

			} else {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to locate CRL").WithCause(err))
			}
			return
		}

		block, _ := pem.Decode([]byte(m.Pem))

		metrics.IncrCounter(mkAIADownloadCrlSuccessful, 1,
			metrics.Tag{Name: ikidTag, Value: ikid},
		)

		w.Header().Set(header.ContentType, "application/pkix-crl")
		w.Write(block.Bytes)
	}
}

// GetCertHandler returns certificate
func (s *Service) GetCertHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
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
			if db.IsNotFoundError(err) {
				// metrics for Not Found
				metrics.IncrCounter(mkAIADownloadCertFailed, 1,
					metrics.Tag{Name: skidTag, Value: skid},
				)
				marshal.WriteJSON(w, r, httperror.WithNotFound("unable to locate certificate"))

			} else {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to locate certificate").WithCause(err))
			}
			return
		}

		block, _ := pem.Decode([]byte(m.Pem))
		metrics.IncrCounter(mkAIADownloadCertSuccessful, 1,
			metrics.Tag{Name: skidTag, Value: skid},
		)

		w.Header().Set(header.ContentType, "application/pkix-cert")
		w.Write(block.Bytes)
	}
}
