package martini

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/ekspand/trusty/pkg/payment"
	"github.com/ekspand/trusty/pkg/poller"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// CreateSubsciptionHandler creates subscription
func (s *Service) CreateSubsciptionHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		req := new(v1.CreateSubscriptionRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		logger.KV(xlog.INFO,
			"user_id", userID,
			"org_id", req.OrgID,
			"product_id", req.ProductID,
		)

		user, err := s.db.GetUser(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("user not found").WithCause(err))
			return
		}

		orgID, err := db.ID(req.OrgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id: "+req.OrgID).WithCause(err))
			return
		}

		ctx := r.Context()
		org, err := s.db.GetOrg(ctx, orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("unable to find organization").WithCause(err))
			return
		}

		subscription, clientSecret, err := s.createSubscription(ctx, user, org, req.ProductID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to create subscription for org %d and product %q", org.ID, req.ProductID).WithCause(err))
			return
		}

		org.Status = v1.OrgStatusPaymentProcessing
		org, err = s.db.UpdateOrgStatus(ctx, org)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to create subscription, status updated failed for org %d", org.ID).WithCause(err))
			return
		}

		res := &v1.CreateSubscriptionResponse{
			Subscription: v1.Subscription{
				OrgID:     db.IDString(subscription.ID),
				Status:    org.Status,
				CreatedAt: subscription.CreatedAt,
				ExpiresAt: subscription.ExpiresAt,
				Price:     subscription.PriceAmount,
				Currency:  subscription.PriceCurrency,
			},
			ClientSecret: clientSecret,
		}
		marshal.WriteJSON(w, r, res)
	}
}

// CancelSubsciptionHandler cancels subscription
func (s *Service) CancelSubsciptionHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {

		req := new(v1.CancelSubscriptionRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		logger.KV(xlog.INFO,
			"user_id", userID,
			"subscription_id", req.SubscriptionID,
		)

		user, err := s.db.GetUser(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("user not found").WithCause(err))
			return
		}

		ctx := r.Context()
		subID, err := db.ID(req.SubscriptionID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to cancel subscription %s", req.SubscriptionID).WithCause(err))
			return
		}
		sub, err := s.db.GetSubscription(ctx, subID, user.ID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to cancel subscription %s", req.SubscriptionID).WithCause(err))
			return
		}

		_, err = s.paymentProv.CancelSubscription(sub.ExternalID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to cancel subscription %s", req.SubscriptionID).WithCause(err))
			return
		}

		org, err := s.db.GetOrg(ctx, sub.ID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("failed to cancel subscription %s", req.SubscriptionID).WithCause(err))
			return
		}
		org.Status = v1.OrgStatusDeactivated

		_, _, err = s.db.UpdateSubscriptionAndOrgStatus(ctx, sub, org)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to cancel subscription").WithCause(err))
			return
		}

		res := &v1.CancelSubscriptionResponse{
			SubscriptionID: db.IDString(sub.ID),
		}
		marshal.WriteJSON(w, r, res)
	}
}

// ListSubsciptionsHandler lists subscriptions
func (s *Service) ListSubsciptionsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		logger.KV(xlog.INFO, "user_id", userID)

		user, err := s.db.GetUser(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("user not found").WithCause(err))
			return
		}

		ctx := r.Context()
		subscriptions, err := s.db.ListSubscriptions(ctx, user.ID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to list subscriptions for user %d", user.ID).WithCause(err))
			return
		}

		res := &v1.ListSubscriptionsResponse{
			Subscriptions: []v1.Subscription{},
		}
		for _, s := range subscriptions {
			res.Subscriptions = append(res.Subscriptions, v1.Subscription{
				OrgID:     db.IDString(s.ID),
				Status:    s.Status,
				CreatedAt: s.CreatedAt,
				ExpiresAt: s.ExpiresAt,
				Price:     s.PriceAmount,
				Currency:  s.PriceCurrency,
			})
		}

		marshal.WriteJSON(w, r, res)
	}
}

// SubscriptionsProductsHandler handles call to list available products for subscriptions
func (s *Service) SubscriptionsProductsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		products := s.paymentProv.ListProducts()
		if len(products) == 0 {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to list any products"))
			return
		}

		res := &v1.SubscriptionsProductsResponse{
			Products: []v1.Product{},
		}
		for _, p := range products {
			res.Products = append(res.Products, v1.Product{
				ID:       p.ID,
				Name:     p.Name,
				Price:    uint64(p.PriceAmount),
				Currency: p.PriceCurrency,
				Years:    uint64(p.Years),
			})
		}
		marshal.WriteJSON(w, r, res)
	}
}

// StripeWebhookHandler handles Stripe webhook calls
func (s *Service) StripeWebhookHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid request").WithCause(err))
			return
		}

		paymentIntent, err := s.paymentProv.HandleWebhook(b, r.Header.Get("Stripe-Signature"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("webhook: invalid request").WithCause(err))
			return
		}

		ctx := r.Context()
		if paymentIntent != nil {
			sub, err := s.db.GetSubscriptionByExternalID(ctx, paymentIntent.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("webhook: unable to find subscription: %s", paymentIntent.ID).WithCause(err))
				return
			}

			sub.Status = paymentIntent.Status

			org, err := s.db.GetOrg(ctx, sub.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("webhook: unable to get org by subscription: %d", sub.ID).WithCause(err))
				return
			}

			if org.Status != v1.OrgStatusApproved {
				org.Status = v1.OrgStatusPaid
			}

			_, _, err = s.db.UpdateSubscriptionAndOrgStatus(ctx, sub, org)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("webhook: unable to update org %d", org.ID).WithCause(err))
				return
			}

		}
		res := &v1.StripeWebhookResponse{}
		marshal.WriteJSON(w, r, res)
	}
}

// createSubscription creates subscription for given user and price in Stripe
func (s *Service) createSubscription(
	ctx context.Context,
	user *model.User,
	org *model.Organization,
	productID string,
) (
	*model.Subscription,
	string,
	error,
) {
	customer, err := s.paymentProv.CreateCustomer(user.Name, user.Email, nil)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription for user name %s, email %s", user.Name, user.Email)
	}

	product, err := s.paymentProv.GetProduct(productID)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to get product with id %s", productID)
	}

	paymentIntent, err := s.paymentProv.CreatePaymentIntent(customer.ID, product.PriceAmount)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription for customer %s, price %s", customer.ID, product.PriceID)
	}

	expiryPeriodInYears := subscriptionExpiryPeriodFromProductName(product.Name)

	now := time.Now().UTC()
	subscriptionModel, err := s.db.CreateSubscription(ctx, &model.Subscription{
		ID:            org.ID,
		ExternalID:    paymentIntent.ID,
		UserID:        user.ID,
		CustomerID:    customer.ID,
		PriceID:       product.PriceID,
		PriceAmount:   uint64(product.PriceAmount),
		PriceCurrency: product.PriceCurrency,
		CreatedAt:     now,
		ExpiresAt:     now.AddDate(expiryPeriodInYears, 0, 0).UTC(),
		Status:        paymentIntent.Status,
	})
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription in db for id %d, customer %s, price %s", org.ID, customer.ID, product.PriceID)
	}

	s.onSubscriptionCreated(subscriptionModel)

	return subscriptionModel, paymentIntent.ClientSecret, nil
}

// onSubscriptionCreated is called when a subscription is created
// it checks and updates payment status
func (s *Service) onSubscriptionCreated(
	sub *model.Subscription,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.Martini.PollPaymentStatusTimeout))

	p := poller.New(nil,
		func(ctx context.Context) (interface{}, error) {
			logger.KV(xlog.TRACE,
				"sub_id", sub.ID,
				"sub_external_id", sub.ExternalID,
				"status", "polling in progress",
			)

			pi, err := s.paymentProv.GetPaymentIntent(sub.ExternalID)
			if err != nil {
				return nil, errors.Annotatef(err, "payment: unable to get payment intent with id %q", sub.ExternalID)
			}

			logger.KV(xlog.TRACE,
				"sub_id", sub.ID,
				"sub_external_id", sub.ExternalID,
				"payment_status", pi.Status,
			)

			sub, err := s.db.GetSubscriptionByExternalID(ctx, pi.ID)
			if err != nil {
				return nil, errors.Annotatef(err, "payment: unable to get subscription with external id %q", pi.ID)
			}

			org, err := s.db.GetOrg(ctx, sub.ID)
			if err != nil {
				return nil, errors.Annotatef(err, "payment: unable to get org with id %q", sub.ID)
			}

			orgStatus := org.Status
			if pi.Status == payment.StatusSucceeded {
				orgStatus = v1.OrgStatusPaid
			}

			if org.Status == orgStatus {
				return orgStatus, nil
			}

			logger.KV(xlog.TRACE,
				"sub_id", sub.ID,
				"sub_external_id", sub.ExternalID,
				"org_current_status", org.Status,
				"org_update_status", orgStatus,
			)

			org.Status = orgStatus
			_, _, err = s.db.UpdateSubscriptionAndOrgStatus(ctx, sub, org)
			if err != nil {
				return "", errors.Annotatef(err, "payment: unable to update status for subscription %d and org %d", sub.ID, org.ID)
			}

			if org.Status == v1.OrgStatusPaid {
				cancel()
			}
			return nil, nil
		},
		func(err error) {
			logger.KV(xlog.ERROR,
				"sub_id", sub.ID,
				"sub_external_id", sub.ExternalID,
				"status", "polling payment status failed",
				"err", errors.Details(err))
		})
	p.Start(ctx, time.Duration(s.cfg.Martini.PollPaymentStatusInterval))

	return nil
}

// subscriptionExpiryPeriodFromProductName derives subscriptions expiry
// period in years for the given product name
func subscriptionExpiryPeriodFromProductName(
	productName string,
) int {
	pName := strings.ToLower(productName)
	if strings.Contains(pName, "1") {
		return 1
	}

	if strings.Contains(pName, "2") {
		return 2
	}

	if strings.Contains(pName, "3") {
		return 3
	}

	if strings.Contains(pName, "4") {
		return 4
	}

	if strings.Contains(pName, "5") {
		return 5
	}

	return 1
}
