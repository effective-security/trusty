package pgsql_test

import (
	"testing"

	"github.com/effective-security/xpki/certutil"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterIssuer(t *testing.T) {
	m := &model.Issuer{
		Label:  certutil.RandomString(32),
		Status: int(pb.IssuerStatus_ACTIVE),
		Config: "# config",
	}

	m1, err := provider.RegisterIssuer(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m)
	defer provider.DeleteIssuer(ctx, m1.Label)

	assert.NotEmpty(t, m1.ID)
	assert.Equal(t, m.Label, m1.Label)
	assert.Equal(t, m.Status, m1.Status)
	assert.Equal(t, m.Config, m1.Config)
	assert.False(t, m1.CreatedAt.IsZero())
	assert.False(t, m1.UpdatedAt.IsZero())

	m2, err := provider.RegisterIssuer(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m2)
	assert.NotEqual(t, *m1, *m2)
	m1.UpdatedAt = m2.UpdatedAt
	assert.Equal(t, *m1, *m2)

	m.Config += " modified"
	m3, err := provider.RegisterIssuer(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m3)
	assert.NotEqual(t, *m2, *m3)
	assert.Equal(t, m.Config, m3.Config)

	m3, err = provider.UpdateIssuerStatus(ctx, m3.ID, int(pb.IssuerStatus_ARCHIVED))
	require.NoError(t, err)
	assert.Equal(t, int(pb.IssuerStatus_ARCHIVED), m3.Status)

	list, err := provider.ListIssuers(ctx, 100, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	err = provider.DeleteIssuer(ctx, m1.Label)
	require.NoError(t, err)
}
