package pgsql_test

import (
	"testing"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterCertProfile(t *testing.T) {
	wm := &model.CertProfile{
		Label:       "wildcard",
		IssuerLabel: "*",
		Config:      "# config",
	}

	mw, err := provider.RegisterCertProfile(ctx, wm)
	require.NoError(t, err)
	require.NotNil(t, mw)
	_ = provider.DeleteCertProfile(ctx, mw.Label)

	issuer := certutil.RandomString(32)
	m := &model.CertProfile{
		Label:       certutil.RandomString(32),
		IssuerLabel: issuer,
		Config:      "# config",
	}

	m1, err := provider.RegisterCertProfile(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m)
	_ = provider.DeleteCertProfile(ctx, m1.Label)

	assert.NotEmpty(t, m1.ID)
	assert.Equal(t, m.Label, m1.Label)
	assert.Equal(t, m.IssuerLabel, m1.IssuerLabel)
	assert.Equal(t, m.Config, m1.Config)
	assert.False(t, m1.CreatedAt.IsZero())
	assert.False(t, m1.UpdatedAt.IsZero())

	m2, err := provider.RegisterCertProfile(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m2)
	assert.Equal(t, *m1, *m2)

	count := 10
	for i := 1; i < count; i++ {
		m := &model.CertProfile{
			Label:       certutil.RandomString(32),
			IssuerLabel: issuer,
			Config:      "# config",
		}

		m1, err := provider.RegisterCertProfile(ctx, m)
		require.NoError(t, err)
		require.NotNil(t, m)
		_ = provider.DeleteCertProfile(ctx, m1.Label)
	}

	list, err := provider.ListCertProfiles(ctx, 100, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	list2, err := provider.GetCertProfilesByIssuer(ctx, issuer)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list2), count+1)

	err = provider.DeleteCertProfile(ctx, m1.Label)
	require.NoError(t, err)
}
