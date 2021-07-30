package acme

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
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

		// check for supported identifiers
		for _, ident := range req.Identifiers {
			if ident.Type != v2acme.IdentifierTNAuthList {
				s.writeProblem(w, r,
					v2acme.MalformedError("NewOrder request included unsupported type identifier: type %q, value %q",
						ident.Type, ident.Value))
				return
			}
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
				//	s.writeProblem(w, r, v2acme.MalformedError("NotBefore and NotAfter are not supported"))
				return
			}
			notBefore = notBefore.UTC()
		}

		// default 90 days
		notAfter := time.Now().Add(2160 * time.Hour).UTC()
		if req.NotAfter != "" {
			notAfter, err = time.Parse("2006-01-02T15:04:05Z", req.NotAfter)
			if err != nil {
				//	s.writeProblem(w, r, v2acme.MalformedError("NotBefore and NotAfter are not supported"))
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
		} else {
			/* TODO: Audit
			s.server.Audit(
				ServiceName,
				evtOrderCreated,
				certcentralID,
				"",
				0,
				fmt.Sprintf("acctID=%s, orderID=%s, names=[%s]",
					account.ID, order.ID, strings.Join(dnsNames, ",")),
			)
			*/
		}

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
