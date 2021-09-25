package acme

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/acme/model"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/martinisecurity/trusty/backend/db"
)

// SetRegistration registers account
func (d *Provider) SetRegistration(ctx context.Context, reg *model.Registration) (*model.Registration, error) {
	return d.db.SetRegistration(ctx, reg)
}

// GetRegistration returns account registration
func (d *Provider) GetRegistration(ctx context.Context, id uint64) (reg *model.Registration, err error) {
	return d.db.GetRegistration(ctx, id)
}

// GetRegistrationByKeyID returns account registration
func (d *Provider) GetRegistrationByKeyID(ctx context.Context, keyID string) (*model.Registration, error) {
	return d.db.GetRegistrationByKeyID(ctx, keyID)
}

// GetOrder returns Order
func (d *Provider) GetOrder(ctx context.Context, id uint64) (order *model.Order, err error) {
	return d.db.GetOrder(ctx, id)
}

// GetOrderByHash returns Order by hash of domain names
func (d *Provider) GetOrderByHash(ctx context.Context, registrationID uint64, namesHash string) (order *model.Order, err error) {
	return d.db.GetOrderByHash(ctx, registrationID, namesHash)
}

// GetOrders returns all Orders for specified registration
func (d *Provider) GetOrders(ctx context.Context, regID uint64) ([]*model.Order, error) {
	return d.db.GetOrders(ctx, regID)
}

// GetAuthorization returns Authorization by ID
func (d *Provider) GetAuthorization(ctx context.Context, authzID uint64) (*model.Authorization, error) {
	return d.db.GetAuthorization(ctx, authzID)
}

// GetAuthorizations returns all Authorizations for specified registration
func (d *Provider) GetAuthorizations(ctx context.Context, regID uint64) ([]*model.Authorization, error) {
	return d.db.GetAuthorizations(ctx, regID)
}

// UpdateOrder updates order
func (d *Provider) UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	return d.db.UpdateOrder(ctx, order)
}

// GetIssuedCertificate returns IssuedCertificate by ID
func (d *Provider) GetIssuedCertificate(ctx context.Context, certID uint64) (*model.IssuedCertificate, error) {
	return d.db.GetIssuedCertificate(ctx, certID)
}

// PutIssuedCertificate saves issued cert
func (d *Provider) PutIssuedCertificate(ctx context.Context, cert *model.IssuedCertificate) (*model.IssuedCertificate, error) {
	return d.db.PutIssuedCertificate(ctx, cert)
}

func (d *Provider) validateOrder(p *model.OrderRequest) error {
	// for IdentifierTNAuthList, only one is supported
	if p.HasIdentifier(v2acme.IdentifierTNAuthList) && len(p.Identifiers) != 1 {
		return errors.Errorf("invalid request: identifiers")
	}

	// TODO: add support for DNS, add config for supported Identifiers
	for _, ident := range p.Identifiers {
		if ident.Type != v2acme.IdentifierTNAuthList {
			return errors.Errorf("NewOrder request included unsupported type identifier: type %q, value %q",
				ident.Type, ident.Value)
		}
	}

	// TODO:
	// Validate that our policy allows issuing for each of the names in the order
	/*
		for _, name := range p.DNSNames {
			if err := d.policy.ValidateDomain(name); err != nil {
				return nil, false, errors.Annotatef(err, "failed validation: %s", name)
			}
		}
	*/

	return nil
}

// NewOrder creates new Order
func (d *Provider) NewOrder(ctx context.Context, p *model.OrderRequest) (*model.Order, bool, error) {
	err := d.validateOrder(p)
	if err != nil {
		return nil, false, errors.Trace(err)
	}

	orderID, err := model.GetIDFromIdentifiers(p.Identifiers)
	if err != nil {
		return nil, false, errors.Trace(err)
	}

	now := time.Now().UTC()
	expiry := d.cfg.Policy.OrderExpiry

	// check for existing order
	order, err := d.GetOrderByHash(ctx, p.RegistrationID, orderID)
	if err == nil {
		// check if update is needed
		// https://datatracker.ietf.org/doc/html/rfc8555#section-7.1.6
		order, err := d.UpdateOrderStatus(ctx, order)
		if err == nil &&
			order.Status.IsPending() &&
			!order.ExpiresAt.IsZero() &&
			order.ExpiresAt.Before(p.NotAfter) &&
			now.Before(order.ExpiresAt) {
			logger.Tracef("reason=found, regID=%d, orderID=%s, status=%v, expires=%s, not_after=%s, now=%s",
				p.RegistrationID, orderID, order.Status,
				order.ExpiresAt.Format(time.RFC3339),
				p.NotAfter.Format(time.RFC3339),
				now.Format(time.RFC3339))
			return order, true, nil
		}
	} else if !db.IsNotFoundError(err) {
		logger.Errorf("err=[%v]", errors.ErrorStack(err))
		// TODO? error or not?
		// return nil, false, errors.Trace(err)
	}

	// TODO: check limit per account

	order = &model.Order{
		NamesHash:         orderID,
		RegistrationID:    p.RegistrationID,
		ExternalBindingID: p.ExternalBindingID,
		NotBefore:         p.NotBefore,
		NotAfter:          p.NotAfter,
		Identifiers:       p.Identifiers,
		CreatedAt:         now,
		Status:            v2acme.StatusPending,
		ExpiresAt:         now.Add(expiry),
	}

	// An order's lifetime is effectively bound by the shortest remaining lifetime
	// of its associated authorizations.
	// To prevent this we only return authorizations that are at least 1 day away
	// from expiring.
	authzExpiryCutoff := now.AddDate(0, 0, 1)

	allAuthorizations, err := d.GetAuthorizations(ctx, p.RegistrationID)
	if err != nil {
		return nil, false, errors.Trace(err)
	}

	// Collect up the authorizations we found into a map keyed by the domains the
	// authorizations correspond to
	nameToExistingAuthz := map[string]*model.Authorization{}
	for _, authz := range allAuthorizations {
		if authz.Status != v2acme.StatusValid && authz.Status != v2acme.StatusPending {
			// skip on invalid status
			continue
		}
		if authz.ExpiresAt.After(authzExpiryCutoff) {
			// skip expired
			continue
		}

		if v2acme.FindIdentifier(p.Identifiers, authz.Identifier.Value) != -1 {
			nameToExistingAuthz[authz.Identifier.Value] = authz
		}
	}

	// Start with the order's own expiry as the minExpiry. We only care
	// about authz expiries that are sooner than the order's expiry
	minExpiry := order.ExpiresAt

	// For each of the names in the order, if there is an acceptable
	// existing authz, append it to the order to reuse it. Otherwise track
	// that there is a missing authz for that name.
	var missingAuthzNames []v2acme.Identifier
	for _, idn := range p.Identifiers {
		// If there isn't an existing authz, note that its missing and continue
		if _, exists := nameToExistingAuthz[idn.Value]; !exists {
			missingAuthzNames = append(missingAuthzNames, idn)
			continue
		}

		authz := nameToExistingAuthz[idn.Value]

		// If the identifier isn't a wildcard, we can reuse any authz.
		// If the identifier is a wildcard and the existing authz only has one
		// DNS-01 type challenge we can reuse it.
		if !strings.HasPrefix(idn.Value, "*.") ||
			(len(authz.Challenges) == 1 && authz.Challenges[0].Type == model.ChallengeTypeDNS01) {
			order.Authorizations = append(order.Authorizations, authz.ID)

			// An authz without an expiry is an unexpected internal server event
			if authz.ExpiresAt.IsZero() {
				return nil, false, errors.NotValidf("expiry of authz %q", authz.ID)
			}

			// If the reused authorization expires before the minExpiry, it's expiry
			// is the new minExpiry.
			if minExpiry.Before(authz.ExpiresAt) {
				minExpiry = authz.ExpiresAt
			}
			continue
		}

		// Delete the authz from the nameToExistingAuthz map since we are not reusing it.
		delete(nameToExistingAuthz, idn.Value)
		// If we reached this point then the existing authz was not acceptable for
		// reuse and we need to mark the name as requiring a new pending authz
		missingAuthzNames = append(missingAuthzNames, idn)
	}

	// If the order isn't fully authorized we need to check that the client has
	// rate limit room for more pending authorizations
	if len(missingAuthzNames) > 0 {
		// TODO: checkPendingAuthorizationLimit(ctx, *order.RegistrationID); err != nil {
	}

	// Loop through each of the names missing authzs and create a new pending
	// authorization for each.
	for _, name := range missingAuthzNames {
		// TODO: Batch check checkInvalidAuthorizationLimit(ctx, registrationID, name)

		authz, err := d.NewAuthorization(ctx, p.RegistrationID, name)
		if err != nil {
			return nil, false, errors.Trace(err)
		}
		order.Authorizations = append(order.Authorizations, authz.ID)

		// If the reused authorization expires before the minExpiry, it's expiry
		// is the new minExpiry.
		if minExpiry.Before(authz.ExpiresAt) {
			minExpiry = authz.ExpiresAt
		}
	}

	order.ExpiresAt = minExpiry

	order, err = d.db.UpdateOrder(ctx, order)
	if err != nil {
		return nil, false, errors.Trace(err)
	}

	return order, false, nil
}

// NewAuthorization creates new Authorization in "pending" state
func (d *Provider) NewAuthorization(ctx context.Context, registrationID uint64, idn v2acme.Identifier) (*model.Authorization, error) {
	expiresAt := time.Now().UTC().Add(d.cfg.Policy.AuthzExpiry).Truncate(time.Second)
	id, _ := d.db.NextID()
	authz := &model.Authorization{
		ID:             id,
		RegistrationID: registrationID,
		Status:         v2acme.StatusPending,
		Identifier:     idn,
		ExpiresAt:      expiresAt,
	}

	// TODO: validate AuthorizationRequest
	//if !d.va.WillingToIssue()

	// TODO: checkPendingAuthorizationLimit

	// TODO: checkInvalidAuthorizationLimit

	// check if Domain is GOOD
	/*
		isSafe, err := d.policy.IsSafeDomain(ctx, dnsName)
		if err != nil {
			return nil, v2acme.ServerInternalError("unable to determine if domain was safe").WithSource(err)
		}
		if !isSafe {
			return nil, v2acme.UnauthorizedError(
				"%q was considered an unsafe domain by a third-party API",
				dnsName,
			)
		}

		if policyCfg.GetReuseValidAuthz() {
			// TODO: check for valid authorizations
		}

		if policyCfg.GetReusePendingAuthz() {
			// TODO: check for pending authorizations
		}
	*/

	var err error
	authz.Challenges, err = d.createChallenges(ctx, registrationID, authz.ID, &authz.Identifier)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Check each challenge for sanity.
	for _, challenge := range authz.Challenges {
		if err := challenge.CheckConsistencyForClientOffer(); err != nil {
			return nil, errors.Annotatef(err, "failed sanity check: %+v", challenge)
		}
	}

	authz, err = d.db.InsertAuthorization(ctx, authz)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return authz, nil
}

func (d *Provider) createChallenges(ctx context.Context, regID, authzID uint64, identifier *v2acme.Identifier) ([]model.Challenge, error) {
	challenges := []model.Challenge{}

	// If the identifier is for a DNS wildcard name we only
	// provide a DNS-01 challenge as a matter of CA policy.
	if identifier.Type == v2acme.IdentifierDNS && strings.HasPrefix(identifier.Value, "*.") {
		// We must have the DNS-01 challenge type enabled to create challenges for
		// a wildcard identifier per LE policy.
		if !d.ChallengeTypeEnabled(model.ChallengeTypeDNS01, regID) {
			return nil, errors.Errorf(
				"challenges requested for wildcard, but DNS-01 challenge type is not enabled")
		}
		id, err := d.db.NextID()
		if err != nil {
			return nil, errors.Trace(err)
		}
		// Only provide a DNS-01-Wildcard challenge
		challenges = []model.Challenge{*model.NewChallenge(id, authzID, model.ChallengeTypeDNS01)}
	} else {
		// Otherwise we collect up challenges based on what is enabled.
		for _, chall := range d.cfg.Policy.EnabledChallenges {
			if d.ChallengeTypeEnabled(chall, regID) {
				id, err := d.db.NextID()
				if err != nil {
					return nil, errors.Trace(err)
				}
				chall := model.NewChallenge(id, authzID, v2acme.IdentifierType(chall))
				challenges = append(challenges, *chall)
			}
		}
	}

	rand.Shuffle(len(challenges), func(i, j int) {
		challenges[i], challenges[j] = challenges[j], challenges[i]
	})

	return challenges, nil
}

// UpdateOrderStatus updates "pending" status depending on aggregate Authorization's status
func (d *Provider) UpdateOrderStatus(ctx context.Context, order *model.Order) (*model.Order, error) {
	status, err := d.statusForOrder(ctx, order)

	if status != order.Status {
		order.Status = status

		order, err = d.UpdateOrder(ctx, order)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return order, nil
}

// GetOrderAuthorizations returns all Authorizations for specified order
func (d *Provider) GetOrderAuthorizations(ctx context.Context, order *model.Order) ([]*model.Authorization, error) {
	list, err := d.GetAuthorizations(ctx, order.RegistrationID)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := make([]*model.Authorization, 0, len(list))
	for _, auth := range list {
		for _, oa := range order.Authorizations {
			if oa == auth.ID {
				res = append(res, auth)
				break
			}
		}
	}

	return res, nil
}

// UpdatePendingAuthorization will update pending Authorization and all its Challenge objects
func (d *Provider) UpdatePendingAuthorization(ctx context.Context, authz *model.Authorization) (*model.Authorization, error) {
	if !authz.Status.IsPending() {
		return nil, errors.Errorf("authorization %d is not in pending state: %s", authz.ID, authz.Status)
	}

	authz, err := d.db.UpdateAuthorization(ctx, authz)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return authz, nil
}

// UpdateAuthorizationAfterValidation will update Authorization and all its Challenge objects after validation
func (d *Provider) UpdateAuthorizationAfterValidation(ctx context.Context, authz *model.Authorization) (*model.Authorization, error) {
	if !authz.Status.IsPending() {
		return nil, errors.Errorf("authorization %d is not in pending state: %s", authz.ID, authz.Status)
	}

	// Consider validation successful if any of the challenges
	// specified in the authorization has been fulfilled
	authz.Status = v2acme.StatusInvalid
	for _, chall := range authz.Challenges {
		if chall.Status == v2acme.StatusValid {
			authz.Status = v2acme.StatusValid
			break
		}
	}

	if authz.Status == v2acme.StatusValid {
		authz.ExpiresAt = time.Now().UTC().Add(d.cfg.Policy.AuthzExpiry).Truncate(time.Second)
	}

	authz, err := d.db.UpdateAuthorization(ctx, authz)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return authz, nil
}

func (d *Provider) statusForOrder(ctx context.Context, order *model.Order) (v2acme.Status, error) {
	// Without any further work we know an order with an error is invalid
	if order.Error != nil || order.Status == v2acme.StatusInvalid {
		return v2acme.StatusInvalid, nil
	}

	now := time.Now().UTC()

	// If the order is expired the status is invalid and we don't need to get
	// order authorizations.
	if order.ExpiresAt.Before(now) {
		return v2acme.StatusInvalid, nil
	}

	// it's already valid and the cert was issued
	if order.Status == v2acme.StatusValid && order.CertificateID != 0 {
		return v2acme.StatusValid, nil
	}

	authzs, err := d.GetOrderAuthorizations(ctx, order)
	if err != nil {
		return v2acme.StatusUnknown, errors.Annotatef(err, "unable to get authorizations")
	}

	// If GetOrderAuthorizations returned a different number of authorization
	// objects than the order's slice of authorization IDs something has gone
	// wrong worth raising an internal error about.
	if len(authzs) != len(order.Authorizations) {
		return v2acme.StatusUnknown, errors.Errorf("unexpected # of auhtorizations: %d vs %d", len(authzs), len(order.Authorizations))
	}

	// Keep a count of the authorizations seen
	invalidAuthzs := 0
	expiredAuthzs := 0
	deactivatedAuthzs := 0
	pendingAuthzs := 0
	validAuthzs := 0

	// Loop over each of the order's authorization objects to examine the authz status
	for _, authz := range authzs {
		switch authz.Status {
		case v2acme.StatusInvalid:
			invalidAuthzs++
		case v2acme.StatusDeactivated:
			deactivatedAuthzs++
		case v2acme.StatusPending:
			pendingAuthzs++
		case v2acme.StatusValid:
			validAuthzs++
		default:
			return v2acme.StatusUnknown, errors.Errorf("auhtorization %d has invalid status %q", authz.ID, authz.Status)
		}

		if authz.ExpiresAt.Before(now) {
			expiredAuthzs++
		}
	}

	logger.Tracef("regID=%d, orderID=%d, reason=authz_count, invalid=%d, deactivated=%d, pending=%d, expired=%d, valid=%d",
		order.RegistrationID, order.ID, invalidAuthzs, deactivatedAuthzs, pendingAuthzs, expiredAuthzs, validAuthzs)

	// An order is invalid if **any** of its authzs are invalid
	if invalidAuthzs > 0 {
		return v2acme.StatusInvalid, nil
	}
	// An order is invalid if **any** of its authzs are expired
	if expiredAuthzs > 0 {
		return v2acme.StatusInvalid, nil
	}
	// An order is deactivated if **any** of its authzs are deactivated
	if deactivatedAuthzs > 0 {
		return v2acme.StatusDeactivated, nil
	}
	// An order is pending if **any** of its authzs are pending
	if pendingAuthzs > 0 {
		return v2acme.StatusPending, nil
	}

	// An order is fully authorized if it has valid authzs for each of the order
	// names
	fullyAuthorized := len(order.Identifiers) == validAuthzs

	// If the order isn't fully authorized we've encountered an internal error:
	// Above we checked for any invalid or pending authzs and should have returned
	// early. Somehow we made it this far but also don't have the correct number
	// of valid authzs.
	if !fullyAuthorized {
		return v2acme.StatusUnknown, errors.Errorf("order %q has inconsistent number of valid authorizations", order.ID)
	}

	if order.Status == v2acme.StatusProcessing {
		return v2acme.StatusProcessing, nil
	}

	return v2acme.StatusReady, nil
}
