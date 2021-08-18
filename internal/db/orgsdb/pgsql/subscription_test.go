package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscription(t *testing.T) {
	now := time.Now()

	userID, err := provider.NextID()
	require.NoError(t, err)
	externalID := certutil.RandomString(32)
	priceID := certutil.RandomString(32)
	customerID := certutil.RandomString(32)
	paymentMethodID := certutil.RandomString(32)

	login1 := fmt.Sprintf("user1%d", userID)
	login2 := fmt.Sprintf("user2%d", userID)
	email1 := fmt.Sprintf("test1%d@ekspand.com", userID)
	email2 := fmt.Sprintf("test2%d@ekspand.com", userID)

	user1, err := provider.LoginUser(ctx, &model.User{
		Login:      login1,
		Email:      email1,
		Name:       email1,
		ExternalID: fmt.Sprintf("%d", userID+1),
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)
	assert.NotNil(t, user1)
	assert.Equal(t, email1, user1.Name)
	assert.Equal(t, email1, user1.Email)
	assert.Equal(t, 1, user1.LoginCount)

	user2, err := provider.LoginUser(ctx, &model.User{
		Login:      login2,
		Email:      email2,
		Name:       email2,
		ExternalID: fmt.Sprintf("%d", userID+2),
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)
	assert.NotNil(t, user2)
	assert.Equal(t, email2, user2.Name)
	assert.Equal(t, email2, user2.Email)
	assert.Equal(t, 1, user2.LoginCount)

	o1, err := provider.UpdateOrg(ctx, &model.Organization{
		Provider:   v1.ProviderGoogle,
		Login:      user1.Login,
		ExternalID: certutil.RandomString(32),
		CreatedAt:  now.UTC(),
		ExpiresAt:  now.UTC().AddDate(2, 0, 0),
		Status:     "incomplete_org",
	})
	require.NoError(t, err)

	m := &model.Subscription{
		ID:              o1.ID,
		ExternalID:      externalID,
		UserID:          userID,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
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
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.Status, m1.Status)

	m1.Status = "approved"
	o1.Status = "org_approved"

	m_upd, o_upd, err := provider.UpdateSubscriptionAndOrgStatus(ctx, m1, o1)
	require.NoError(t, err)
	require.NotNil(t, m_upd)
	require.NotNil(t, o_upd)

	assert.Equal(t, m1.ID, m_upd.ID)
	assert.Equal(t, "approved", m_upd.Status)

	assert.Equal(t, o1.ID, o_upd.ID)
	assert.Equal(t, "org_approved", o_upd.Status)

	m4, err := provider.GetSubscription(ctx, m_upd.ID, userID)
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
	assert.Equal(t, m1.CreatedAt.Unix(), m4.CreatedAt.Unix())
	assert.Equal(t, m1.ExpiresAt.Unix(), m4.ExpiresAt.Unix())
	assert.Equal(t, "approved", m4.Status)

	o4, err := provider.GetOrg(ctx, o_upd.ID)
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
	assert.Equal(t, m1.CreatedAt.Unix(), m5.CreatedAt.Unix())
	assert.Equal(t, m1.ExpiresAt.Unix(), m5.ExpiresAt.Unix())
	assert.Equal(t, "approved", m5.Status)

	o2, err := provider.UpdateOrg(ctx, &model.Organization{
		Provider:   v1.ProviderGoogle,
		Login:      user2.Login,
		ExternalID: certutil.RandomString(32),
		CreatedAt:  now.UTC(),
		ExpiresAt:  now.UTC().AddDate(2, 0, 0),
		Status:     "incomplete_org",
	})
	require.NoError(t, err)
	m2 := &model.Subscription{
		ID:              o2.ID,
		ExternalID:      certutil.RandomString(32),
		UserID:          userID,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
		CreatedAt:       now.UTC(),
		ExpiresAt:       now.UTC().AddDate(2, 0, 0),
		Status:          "incomplete",
	}
	_, err = provider.CreateSubscription(ctx, m2)
	require.NoError(t, err)

	o3, err := provider.UpdateOrg(ctx, &model.Organization{
		Provider:   v1.ProviderGithub,
		Login:      user1.Login,
		ExternalID: certutil.RandomString(32),
		CreatedAt:  now.UTC(),
		ExpiresAt:  now.UTC().AddDate(2, 0, 0),
		Status:     "incomplete_org",
	})
	require.NoError(t, err)

	m3 := &model.Subscription{
		ID:              o3.ID,
		ExternalID:      certutil.RandomString(32),
		UserID:          userID,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
		CreatedAt:       now.UTC(),
		ExpiresAt:       now.UTC().AddDate(2, 0, 0),
		Status:          "incomplete",
	}
	_, err = provider.CreateSubscription(ctx, m3)
	require.NoError(t, err)

	userID3, err := provider.NextID()
	require.NoError(t, err)
	_, err = provider.CreateSubscription(ctx, &model.Subscription{
		ID:              o1.ID,
		ExternalID:      certutil.RandomString(32),
		UserID:          userID3,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
		CreatedAt:       now.UTC(),
		ExpiresAt:       now.UTC().AddDate(2, 0, 0),
		Status:          "incomplete",
	})
	require.NoError(t, err)

	subs, err := provider.ListSubscriptions(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 3, len(subs))

	require.Equal(t, now.UTC().Unix(), subs[0].CreatedAt.Unix())
	require.Equal(t, now.UTC().Unix(), subs[1].CreatedAt.Unix())
	require.Equal(t, now.UTC().Unix(), subs[2].CreatedAt.Unix())

	userID2, err := provider.NextID()
	require.NoError(t, err)
	subUser2, err := provider.CreateSubscription(ctx, &model.Subscription{
		ID:              o1.ID,
		ExternalID:      certutil.RandomString(32),
		UserID:          userID2,
		CustomerID:      customerID,
		PriceID:         priceID,
		PriceAmount:     12,
		PriceCurrency:   "USD",
		PaymentMethodID: paymentMethodID,
		CreatedAt:       now.UTC(),
		ExpiresAt:       now.UTC().AddDate(2, 0, 0),
		Status:          "incomplete",
	})
	require.NoError(t, err)

	err = provider.RemoveSubscription(ctx, subUser2.ID)
	require.NoError(t, err)
	_, err = provider.GetSubscription(ctx, subUser2.ID, userID2)
	require.Error(t, err)
}
