package acme

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
)

var (
	keyForACMEOrder = []string{"acme", "order"}
)

// NewOrderHandler creates an order
func (s *Service) NewOrderHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)

		ctx := r.Context()
		body, _, account, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		var req v2acme.OrderRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("failed unmarshaling JSON").WithSource(err))
			return
		}

		if len(req.Identifiers) == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("NewOrder request did not specify any identifiers"))
			return
		}

		// check the Org is approved
		orgID, _ := db.ID(account.ExternalID)
		org, err := s.orgsdb.GetOrg(ctx, orgID)
		if err != nil || org.Status != v1.OrgStatusApproved {
			s.writeProblem(w, r, v2acme.MalformedError("organization is not in Approved state"))
			return
		}

		notBefore := time.Now()
		if req.NotBefore != "" {
			notBefore, err = time.Parse("2006-01-02T15:04:05Z", req.NotBefore)
			if err != nil {
				s.writeProblem(w, r, v2acme.MalformedError("invalid not_before value: %s", err.Error()))
				return
			}
			notBefore = notBefore.UTC()
		}

		// default 90 days
		notAfter := time.Now().Add(2160 * time.Hour).UTC()
		if req.NotAfter != "" {
			notAfter, err = time.Parse("2006-01-02T15:04:05Z", req.NotAfter)
			if err != nil {
				s.writeProblem(w, r, v2acme.MalformedError("invalid not_after value: %s", err.Error()))
				return
			}
			notAfter = notAfter.UTC()
		}

		if notAfter.After(org.ExpiresAt) {
			notAfter = org.ExpiresAt
		}
		maxtime := time.Now().Add(8760 * time.Hour).UTC()
		if notAfter.After(maxtime) {
			notAfter = maxtime
		}

		order, existing, err := s.controller.NewOrder(ctx, &acmemodel.OrderRequest{
			RegistrationID:    account.ID,
			ExternalBindingID: account.ExternalID,
			NotBefore:         notBefore,
			NotAfter:          notAfter,
			Identifiers:       req.Identifiers,
		})
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("unable to create order").WithSource(err))
			return
		}

		status := http.StatusCreated
		if existing {
			status = http.StatusOK
			if order.Status == v2acme.StatusPending {
				// check if update is needed
				orderUpdated, err := s.controller.UpdateOrderStatus(ctx, order)
				if err != nil {
					s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive order %d/%d", account.ID, order.ID).WithSource(err))
					return
				}

				order = orderUpdated
			}
		}

		tags := []metrics.Tag{
			{Name: "account", Value: db.IDString(account.ID)},
			{Name: "existing", Value: fmt.Sprintf("%t", existing)},
		}

		metrics.IncrCounter(keyForACMEOrder, 1, tags...)

		s.writeOrder(w, r, status, order)
	}
}

// GetOrderHandler returns order
func (s *Service) GetOrderHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)

		acctID, _ := db.ID(p.ByName("acct_id"))
		orderID, _ := db.ID(p.ByName("id"))

		if acctID == 0 || orderID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d\"", acctID, orderID))
			return
		}

		ctx := r.Context()
		_, _, _, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		order, err := s.controller.GetOrder(ctx, orderID)
		if err != nil {
			// If the order isn't found, return a suitable problem
			if db.IsNotFoundError(err) {
				s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d not found", acctID, orderID))
			} else {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive order %d/%d", acctID, orderID))
			}
			return
		}

		if order.RegistrationID != acctID {
			s.writeProblem(w, r, v2acme.MalformedError("invalid acct_id %d/%d", acctID, orderID))
			return
		}

		now := time.Now().UTC()
		if order.ExpiresAt.IsZero() || order.ExpiresAt.Before(now) {
			s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d has expired: %s",
				acctID, orderID, order.ExpiresAt.Format(time.RFC3339)))
			return
		}

		if order.Status == v2acme.StatusPending {
			// check if update is needed
			orderUpdated, err := s.controller.UpdateOrderStatus(ctx, order)
			if err != nil {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to determine order status %d/%d", acctID, orderID).WithSource(err))
				return
			}

			order = orderUpdated
			w.Header().Set("Retry-After", "10")
		} else if order.Status == v2acme.StatusProcessing {
			/* TODO
			go func(order *acmemodel.Order) {
				_, err := s.checkProcessingOrderStatus(order)
				if err != nil {
					logger.Errorf("api=checkProcessingOrderStatus, certcentralID=%s, regID=%s, orderID=%s, externalID=%d, status=%v, err=[%v]",
						order.ExternalBindingID, order.RegistrationID, order.ID, order.ExternalOrderID, order.Status, err.Error())

				}
			}(order)
			*/
		}

		s.writeOrder(w, r, http.StatusOK, order)
	}
}

// FinalizeOrderHandler handles CSR
func (s *Service) FinalizeOrderHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)
		acctID, _ := db.ID(p.ByName("acct_id"))
		orderID, _ := db.ID(p.ByName("id"))

		if acctID == 0 || orderID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d\"", acctID, orderID))
			return
		}

		ctx := r.Context()
		body, _, account, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		if account.ID != acctID {
			s.writeProblem(w, r, v2acme.UnauthorizedError("user account ID doesn't match account ID in authorization: %q", acctID))
			return
		}

		// The authenticated finalize message body should be an encoded CSR
		var req v2acme.CertificateRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("unable decode finalize order request").WithSource(err))
			return
		}

		csr, err := x509.ParseCertificateRequest(req.CSR)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("unable to parse CSR").WithSource(err))
			return
		}

		order, err := s.controller.GetOrder(ctx, orderID)
		if err != nil {
			// If the order isn't found, return a suitable problem
			if db.IsNotFoundError(err) {
				s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d not found", acctID, orderID))
			} else {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive order %d/%d", acctID, orderID))
			}
			return
		}

		if order.Status == v2acme.StatusInvalid {
			s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d is invalid: %v", acctID, orderID, order.Error))
			return
		}

		now := time.Now().UTC()
		if order.ExpiresAt.IsZero() || order.ExpiresAt.Before(now) {
			s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d has expired: %s",
				acctID, orderID, order.ExpiresAt.Format(time.RFC3339)))
			return
		}

		// TODO: verification
		/*
			// Dedupe, lowercase and sort both the names from the CSR and the names in the
			// order.
			csrNames := acmemodel.UniqueLowerNames(csr.DNSNames)
			orderNames := acmemodel.UniqueLowerNames(order.DNSNames)

			// Immediately reject the request if the number of names differ
			if len(orderNames) != len(csrNames) {
				s.writeProblem(w, r, v2acme.UnauthorizedError("CSR includes different number of names than specifies in the order: [%s]",
					strings.Join(csrNames, ",")))
				return
			}

			// Check that the order names and the CSR names are an exact match
			for i, name := range orderNames {
				if name != csrNames[i] {
					s.writeProblem(w, r, v2acme.UnauthorizedError("CSR is missing Order domain: %q", name))
					return
				}
			}
		*/
		if order.Status == v2acme.StatusPending || order.Status == v2acme.StatusReady {
			// check if update is needed
			order.Status = v2acme.StatusProcessing
			orderUpdated, err := s.controller.UpdateOrderStatus(ctx, order)
			if err != nil {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive order %d/%d", acctID, orderID).WithSource(err))
				return
			}

			order = orderUpdated
		}

		if order.Status != v2acme.StatusProcessing {
			s.writeProblem(w, r, v2acme.NotFoundError("order %d/%d is not ready: %v", acctID, orderID, order.Status))
			return
		}

		// TODO:
		// - request and issue Cert
		// - update order with Cert and StatusValid

		var subject *pb.X509Subject
		if order.HasIdentifier(v2acme.IdentifierTNAuthList) {
			tn, _ := acmemodel.ParseTNEntry(order.Identifiers[0].Value)

			subject = &pb.X509Subject{
				CommonName: "SHAKEN " + tn.SPC.Code,
				Names: []*pb.X509Name{
					{
						Country:      "US",
						Organisation: "Entity Name From Registration",
					},
				},
			}
		} else {
			/*
				cn := csr.Subject.CommonName
				if cn == "" {
					cn = order.DNSNames[0]
				}

				san := order.DNSNames
				if len(san) == 1 && san[0] == cn {
					san = nil
				}
			*/
		}

		// TODO: post to the Queue
		ca, err := s.getCAClient()
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("CA is unavailable %d/%d", acctID, orderID).WithSource(err))
			return
		}

		order.Status = v2acme.StatusProcessing
		order, err = s.controller.UpdateOrder(ctx, order)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("unable to update order %d/%d", acctID, orderID).WithSource(err))
			return
		}

		orgID, _ := db.ID(account.ExternalID)

		signReq := &pb.SignCertificateRequest{
			Profile:       "SHAKEN",
			RequestFormat: pb.EncodingFormat_PEM,
			Request:       []byte(encodeCSR(csr.Raw)),
			Subject:       subject,
			OrgId:         orgID,
		}

		res, err := ca.SignCertificate(ctx, signReq)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("failed to request certificate").WithSource(err))
			return
		}

		certPem := strings.TrimSpace(res.Certificate.Pem) + "\n" + strings.TrimSpace(res.Certificate.IssuersPem)

		logger.KV(xlog.DEBUG, "issued", certPem)

		cert, err := s.controller.PutIssuedCertificate(ctx, &acmemodel.IssuedCertificate{
			ID:             res.Certificate.Id,
			RegistrationID: order.RegistrationID,
			OrderID:        order.ID,
			Certificate:    certPem,
			ExternalID:     res.Certificate.Id,
		})
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("failed to store certificate").WithSource(err))
			return
		}

		order.Status = v2acme.StatusValid
		order.CertificateID = cert.ID
		order, err = s.controller.UpdateOrder(ctx, order)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("unable to update order %d/%d", acctID, orderID).WithSource(err))
			return
		}

		// TODO:
		/*
			s.server.Audit(
				ServiceName,
				evtCSRPosted,
				account.ExternalID,
				"",
				0,
				fmt.Sprintf("acctID=%s, orderID=%s, CN=%q, DNSNames=[%s]",
					account.ID, order.ID, cn, strings.Join(san, ",")),
			)
		*/

		s.writeOrder(w, r, http.StatusOK, order)
	}
}

// GetCertHandler returns certificate
func (s *Service) GetCertHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)
		acctID, _ := db.ID(p.ByName("acct_id"))
		certID, _ := db.ID(p.ByName("id"))

		if acctID == 0 || certID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d\"", acctID, certID))
			return
		}

		ctx := r.Context()
		_, _, account, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		if account.ID != acctID {
			s.writeProblem(w, r, v2acme.UnauthorizedError("user account ID doesn't match account ID in authorization: %q", acctID))
			return
		}

		cert, err := s.controller.GetIssuedCertificate(ctx, certID)
		if err != nil {
			// If the certificate isn't found, return a suitable problem
			if db.IsNotFoundError(err) {
				s.writeProblem(w, r, v2acme.NotFoundError("certificate %d/%d not found", acctID, certID))
			} else {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive certificate %d/%d", acctID, certID))
			}
			return
		}

		w.Header().Set(header.ContentType, "application/pem-certificate-chain")
		w.Write([]byte(cert.Certificate))
	}
}

func (s *Service) writeOrder(w http.ResponseWriter, r *http.Request, statusCode int, order *acmemodel.Order) {
	orderURL := s.baseURL() + fmt.Sprintf(uriOrderByIDFmt, order.RegistrationID, order.ID)
	finalizeURL := s.baseURL() + fmt.Sprintf(uriFinalizeByIDFmt, order.RegistrationID, order.ID)

	// set location header
	// Location: https://xxx.com/v2/acme/account/:acctID/orders/:orderID
	w.Header().Set(header.Location, orderURL)

	o := v2acme.Order{
		Status:         order.Status,
		ExpiresAt:      order.ExpiresAt.Format(time.RFC3339),
		Identifiers:    order.Identifiers,
		NotBefore:      order.NotBefore.UTC().Format(time.RFC3339),
		NotAfter:       order.NotAfter.UTC().Format(time.RFC3339),
		Error:          order.Error,
		Authorizations: make([]string, len(order.Authorizations)),
		FinalizeURL:    finalizeURL,
	}

	for i, authz := range order.Authorizations {
		// https://xxx.com/v2/acme/account/:acctID/authz/:id
		o.Authorizations[i] = s.baseURL() + fmt.Sprintf(uriAuthzByIDFmt, order.RegistrationID, authz)
	}

	if order.Status == v2acme.StatusValid {
		// https://xxx.com/v2/acme/account/:acctID/cert/:certID
		o.CertificateURL = s.baseURL() + fmt.Sprintf(uriCertByIDFmt, order.RegistrationID, order.CertificateID)
	}

	logger.KV(xlog.NOTICE, "order", order)

	marshal.WritePlainJSON(w, statusCode, &o, marshal.PrettyPrint)
}

func encodeCSR(csr []byte) string {
	b := bytes.NewBuffer([]byte{})
	_ = pem.Encode(b, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
	return string(b.Bytes())
}
