package acme

import (
	"context"

	"github.com/ekspand/trusty/acme/model"
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

// GetOrder returns Order by hash of domain names
func (d *Provider) GetOrder(ctx context.Context, registrationID uint64, namesHash string) (order *model.Order, err error) {
	return d.db.GetOrder(ctx, registrationID, namesHash)
}

// GetOrders returns all Orders for specified registration
func (d *Provider) GetOrders(ctx context.Context, regID uint64) ([]*model.Order, error) {
	return d.db.GetOrders(ctx, regID)
}
