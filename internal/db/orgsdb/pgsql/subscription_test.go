package pgsql_test

import (
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscription(t *testing.T) {
	now := time.Now()
	o1, err := provider.UpdateOrg(ctx, &model.Organization{
		ExternalID: certutil.RandomString(32),
		CreatedAt:  now.UTC(),
		ExpiresAt:  now.UTC().AddDate(2, 0, 0),
		Status:     "incomplete_org",
	})
	require.NoError(t, err)

	userID, err := provider.NextID()
	require.NoError(t, err)
	externalID := certutil.RandomString(32)
	priceID := certutil.RandomString(32)
	customerID := certutil.RandomString(32)
	paymentMethodID := certutil.RandomString(32)

	m := &model.Subscription{
		ID:              o1.ID,
		ExternalID:      externalID,
		UserID:          userID,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
		Years:           2,
		CreatedAt:       now.UTC(),
		ExpiresAt:       now.UTC().AddDate(2, 0, 0),
		Status:          "incomplete",
	}

	m1, err := provider.CreateSubscription(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m1)
	assert.Equal(t, m.ID, m1.ID)
	assert.Equal(t, m.ExternalID, m1.ExternalID)
	assert.Equal(t, m.UserID, m1.UserID)
	assert.Equal(t, m.CustomerID, m1.CustomerID)
	assert.Equal(t, m.PriceID, m1.PriceID)
	assert.Equal(t, m.PriceAmount, m1.PriceAmount)
	assert.Equal(t, m.PriceCurrency, m1.PriceCurrency)
	assert.Equal(t, m.PaymentMethodID, m1.PaymentMethodID)
	assert.Equal(t, m.Years, m1.Years)
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.Status, m1.Status)

	// run create again to make sure it does not fail and instead does update
	m1, err = provider.CreateSubscription(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m1)
	assert.Equal(t, m.ID, m1.ID)
	assert.Equal(t, m.ExternalID, m1.ExternalID)
	assert.Equal(t, m.UserID, m1.UserID)
	assert.Equal(t, m.CustomerID, m1.CustomerID)
	assert.Equal(t, m.PriceID, m1.PriceID)
	assert.Equal(t, m.PriceAmount, m1.PriceAmount)
	assert.Equal(t, m.PriceCurrency, m1.PriceCurrency)
	assert.Equal(t, m.PaymentMethodID, m1.PaymentMethodID)
	assert.Equal(t, m.Years, m1.Years)
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.Status, m1.Status)

	m1.Status = "approved"
	o1.Status = "org_approved"

	m3, o3, err := provider.UpdateSubscriptionAndOrgStatus(ctx, m1, o1)
	require.NoError(t, err)
	require.NotNil(t, m3)
	require.NotNil(t, o3)

	assert.Equal(t, m1.ID, m3.ID)
	assert.Equal(t, "approved", m3.Status)

	assert.Equal(t, o1.ID, o3.ID)
	assert.Equal(t, "org_approved", o3.Status)

	m4, err := provider.GetSubscription(ctx, m3.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, m4)
	assert.Equal(t, m1.ID, m4.ID)
	assert.Equal(t, m1.ExternalID, m4.ExternalID)
	assert.Equal(t, m1.UserID, m4.UserID)
	assert.Equal(t, m1.CustomerID, m4.CustomerID)
	assert.Equal(t, m1.PriceID, m4.PriceID)
	assert.Equal(t, m1.PriceAmount, m4.PriceAmount)
	assert.Equal(t, m1.PriceCurrency, m4.PriceCurrency)
	assert.Equal(t, m1.PaymentMethodID, m4.PaymentMethodID)
	assert.Equal(t, m1.Years, m4.Years)
	assert.Equal(t, m1.CreatedAt.Unix(), m4.CreatedAt.Unix())
	assert.Equal(t, m1.ExpiresAt.Unix(), m4.ExpiresAt.Unix())
	assert.Equal(t, "approved", m4.Status)

	o4, err := provider.GetOrg(ctx, o3.ID)
	require.NoError(t, err)
	require.NotNil(t, o4)
	assert.Equal(t, o1.ID, o4.ID)
	assert.Equal(t, o1.ExternalID, o4.ExternalID)
	assert.Equal(t, o1.CreatedAt.Unix(), o4.CreatedAt.Unix())
	assert.Equal(t, o1.ExpiresAt.Unix(), o4.ExpiresAt.Unix())
	assert.Equal(t, "org_approved", o4.Status)

	m5, err := provider.GetSubscriptionByExternalID(ctx, m4.ExternalID)
	require.NoError(t, err)
	require.NotNil(t, m5)
	assert.Equal(t, m1.ID, m5.ID)
	assert.Equal(t, m1.ExternalID, m5.ExternalID)
	assert.Equal(t, m1.UserID, m5.UserID)
	assert.Equal(t, m1.CustomerID, m5.CustomerID)
	assert.Equal(t, m1.PriceID, m5.PriceID)
	assert.Equal(t, m1.PriceAmount, m5.PriceAmount)
	assert.Equal(t, m1.PriceCurrency, m5.PriceCurrency)
	assert.Equal(t, m1.PaymentMethodID, m5.PaymentMethodID)
	assert.Equal(t, m1.Years, m5.Years)
	assert.Equal(t, m1.CreatedAt.Unix(), m5.CreatedAt.Unix())
	assert.Equal(t, m1.ExpiresAt.Unix(), m5.ExpiresAt.Unix())
	assert.Equal(t, "approved", m5.Status)
}
