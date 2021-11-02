package cis

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"golang.org/x/crypto/ocsp"
	"google.golang.org/grpc/codes"
)

var (
	malformedRequestErrorResponse = []byte{0x30, 0x03, 0x0A, 0x01, 0x01}
	internalErrorErrorResponse    = []byte{0x30, 0x03, 0x0A, 0x01, 0x02}
	tryLaterErrorResponse         = []byte{0x30, 0x03, 0x0A, 0x01, 0x03}
	sigRequredErrorResponse       = []byte{0x30, 0x03, 0x0A, 0x01, 0x05}
	unauthorizedErrorResponse     = []byte{0x30, 0x03, 0x0A, 0x01, 0x06}
)

// OCSPResponder represents the logical source of OCSP responses, i.e.,
// the logic that actually chooses a response based on a request.
// In order to create an actual responder, wrap one of these in a Responder
// object and pass it to http.Handle.
// By default the Responder will set the headers:
//
// Cache-Control to "max-age=(response.NextUpdate-now), public, no-transform, must-revalidate",
// Last-Modified to response.ThisUpdate,
// Expires to response.NextUpdate,
// ETag to the SHA256 hash of the response,
// Content-Type to application/ocsp-response.
//
// To override these headers, or set extra headers,
// OCSPResponder should return a http.Header with the headers to override,
// or nil otherwise.
type OCSPResponder interface {
	Response(*ocsp.Request) ([]byte, http.Header, error)
}

// GetOcspHandler returns OCSP via GET
func (s *Service) GetOcspHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		query := p.ByName("body")
		base64Request, err := url.QueryUnescape(query)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to parse query"))
			return
		}

		// url.QueryUnescape not only unescapes %2B escaping, but it additionally
		// turns the resulting '+' into a space, which makes base64 decoding fail.
		// So we go back afterwards and turn ' ' back into '+'. This means we
		// accept some malformed input that includes ' ' or %20, but that's fine.
		base64RequestBytes := []byte(base64Request)
		for i := range base64RequestBytes {
			if base64RequestBytes[i] == ' ' {
				base64RequestBytes[i] = '+'
			}
		}
		// In certain situations a UA may construct a request that has a double
		// slash between the host name and the base64 request body due to naively
		// constructing the request URL. In that case strip the leading slash
		// so that we can still decode the request.
		if len(base64RequestBytes) > 0 && base64RequestBytes[0] == '/' {
			base64RequestBytes = base64RequestBytes[1:]
		}
		requestBody, err := base64.StdEncoding.DecodeString(string(base64RequestBytes))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to parse request"))
			return
		}
		s.ocspResponse(w, r, requestBody)
	}
}

// OcspHandler returns OCSP
func (s *Service) OcspHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to read request"))
			return
		}

		s.ocspResponse(w, r, requestBody)
	}
}

func (s *Service) ocspResponse(w http.ResponseWriter, r *http.Request, requestBody []byte) {
	// All responses after this point will be OCSP.
	// We could check for the content type of the request, but that
	// seems unnecessariliy restrictive.
	w.Header().Add(header.ContentType, "application/ocsp-response")

	// logger.Tracef("api=ocsp, req=%x", requestBody)

	res, err := s.ca.SignOCSP(r.Context(), &pb.OCSPRequest{Der: requestBody})
	if err != nil {
		logger.KV(xlog.WARNING, "err", err.Error())

		if trustyErr, ok := err.(v1.TrustyError); ok {
			switch trustyErr.Code() {
			case codes.InvalidArgument:
				w.Write(malformedRequestErrorResponse)
				return
			case codes.NotFound:
				w.Write(unauthorizedErrorResponse)
				return
			}
		}

		w.Write(internalErrorErrorResponse)
		return
	}

	w.Write(res.Der)
}
