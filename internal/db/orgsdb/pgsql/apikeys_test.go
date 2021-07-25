package pgsql_test

import (
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeys(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	key := certutil.RandomString(32)
	now := time.Now()
	m := &model.APIKey{
		OrgID:      id,
		Key:        key,
		Enrollemnt: true,
		CreatedAt:  now.UTC(),
		ExpiresAt:  now.Add(8700 * time.Hour).UTC(),
	}

	m1, err := provider.CreateAPIKey(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m1)
	defer provider.DeleteApprovalToken(ctx, m1.ID)

	assert.Equal(t, m.OrgID, m1.OrgID)
	assert.Equal(t, m.Key, m1.Key)
	assert.Equal(t, m.Enrollemnt, m1.Enrollemnt)
	assert.Equal(t, m.Management, m1.Management)
	assert.Equal(t, m.Billing, m1.Billing)
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.UsedAt.Unix(), m1.UsedAt.Unix())

	_, err = provider.CreateAPIKey(ctx, m)
	require.Error(t, err)
	assert.Equal(t, "pq: duplicate key value violates unique constraint \"apikeys_key\"", err.Error())

	_, err = provider.GetAPIKey(ctx, "notfound")
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	m2, err := provider.GetAPIKey(ctx, m.Key)
	require.NoError(t, err)
	assert.Equal(t, m.Key, m2.Key)
	assert.Equal(t, m.Enrollemnt, m2.Enrollemnt)
	assert.Equal(t, m.Management, m2.Management)
	assert.Equal(t, m.Billing, m2.Billing)
	assert.Equal(t, m.CreatedAt.Unix(), m2.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m2.ExpiresAt.Unix())
	assert.True(t, m2.UsedAt.After(m.UsedAt))

	list, err := provider.GetOrgAPIKeys(ctx, m.OrgID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}
