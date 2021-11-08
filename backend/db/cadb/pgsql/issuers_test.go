package pgsql_test

import (
	"testing"

	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterIssuer(t *testing.T) {
	m := &model.Issuer{
		Label:  certutil.RandomString(32),
		Config: "# config",
	}

	m1, err := provider.RegisterIssuer(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m)
	defer provider.DeleteIssuer(ctx, m1.Label)

	assert.NotEmpty(t, m1.ID)
	assert.Equal(t, m.Label, m1.Label)
	assert.Equal(t, m.Config, m1.Config)
	assert.False(t, m1.CreatedAt.IsZero())
	assert.False(t, m1.UpdatedAt.IsZero())

	m2, err := provider.RegisterIssuer(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m2)
	assert.Equal(t, *m1, *m2)

	list, err := provider.ListIssuers(ctx, 100, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	err = provider.DeleteIssuer(ctx, m1.Label)
	require.NoError(t, err)
}
