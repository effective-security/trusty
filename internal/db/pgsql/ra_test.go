package pgsql_test

import (
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRootCertificate(t *testing.T) {
	rc := &model.RootCertificate{
		SKID:             guid.MustCreate(),
		NotBefore:        time.Now().Add(-time.Hour).UTC(),
		NotAfter:         time.Now().Add(time.Hour).UTC(),
		ThumbprintSha256: "12345",
		Trust:            1,
		Pem:              "pem",
	}

	r, err := provider.RegisterRootCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer provider.RemoveRootCertificate(ctx, r.ID)

	assert.Equal(t, rc.SKID, r.SKID)
	assert.Equal(t, rc.ThumbprintSha256, r.ThumbprintSha256)
	assert.Equal(t, rc.Trust, r.Trust)
	assert.Equal(t, rc.Pem, r.Pem)
	assert.Equal(t, rc.NotBefore.Unix(), r.NotBefore.Unix())
	assert.Equal(t, rc.NotAfter.Unix(), r.NotAfter.Unix())

	r2, err := provider.RegisterRootCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, *r, *r2)

	list, err := provider.GetRootCertificates(ctx)
	require.NoError(t, err)
	r3 := findRoot(list, r.ID)
	require.NotNil(t, r3)
	assert.Equal(t, *r, *r3)
}

func findRoot(list []*model.RootCertificate, id int64) *model.RootCertificate {
	for _, m := range list {
		if m.ID == id {
			return m
		}
	}
	return nil
}
