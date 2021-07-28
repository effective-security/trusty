package martini

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
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
			"subscription_years", req.SubscriptionYears,
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

		subscription, clientSecret, err := s.createSubscription(ctx, user, org, req.SubscriptionYears)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to process the subscription").WithCause(err))
			return
		}

		org.Status = v1.OrgStatusValidationPending
		_, err = s.db.UpdateOrgStatus(ctx, org)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to process the subscription").WithCause(err))
			return
		}

		res := &v1.CreateSubscriptionResponse{
			OrgID:             strconv.FormatUint(subscription.ID, 10),
			SubscriptionYears: subscription.Years,
			PriceAmount:       subscription.PriceAmount,
			PriceCurrency:     subscription.PriceCurrency,
			Status:            subscription.Status,
			ClientSecret:      clientSecret,
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
			SubscriptionID: strconv.FormatUint(sub.ID, 10),
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

		subscriptionFromStripe, err := s.paymentProv.HandleWebhook(b, r.Header.Get("Stripe-Signature"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid request").WithCause(err))
			return
		}

		ctx := r.Context()
		if subscriptionFromStripe != nil {
			sub, err := s.db.GetSubscriptionByExternalID(ctx, subscriptionFromStripe.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to handle webhook call").WithCause(err))
				return
			}

			org, err := s.db.GetOrg(ctx, sub.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to handle webhook call").WithCause(err))
				return
			}

			if org.Status != v1.OrgStatusApproved {
				org.Status = v1.OrgStatusValidationPending
			}

			_, _, err = s.db.UpdateSubscriptionAndOrgStatus(ctx, sub, org)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to handle webhook call").WithCause(err))
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
	years uint32,
) (
	*model.Subscription,
	string,
	error,
) {
	customer, err := s.paymentProv.CreateCustomer(user.Name, user.Email, nil)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription for user name %s, email %s", user.Name, user.Email)
	}

	product, err := s.paymentProv.GetProduct(years)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to get product for years %d", years)
	}

	subscription, err := s.paymentProv.CreateSubscription(customer.ID, product.PriceID)
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription for customer %s, price %s", customer.ID, product.PriceID)
	}

	now := time.Now().UTC()
	subscriptionModel, err := s.db.CreateSubscription(ctx, &model.Subscription{
		ID:            org.ID,
		ExternalID:    subscription.ID,
		UserID:        user.ID,
		CustomerID:    customer.ID,
		PriceID:       product.PriceID,
		PriceAmount:   uint64(product.PriceAmount),
		PriceCurrency: product.PriceCurrency,
		Years:         years,
		CreatedAt:     now,
		ExpiresAt:     now.AddDate(int(years), 0, 0),
		Status:        subscription.Status,
	})
	if err != nil {
		return nil, "", errors.Annotatef(err, "unable to create a subscription in db for id %d, customer %s, price %s", org.ID, customer.ID, product.PriceID)
	}
	return subscriptionModel, subscription.ClientSecret, nil
}
