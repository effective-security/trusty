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
	id, err := provider.NextID()
	require.NoError(t, err)
	ownerID := int64(id)

	o := &model.RootCertificate{
		OwnerID:          ownerID,
		Skid:             guid.MustCreate(),
		NotBefore:        time.Now().Add(-time.Hour).UTC(),
		NotAfter:         time.Now().Add(time.Hour).UTC(),
		ThumbprintSha256: "12345",
		Trust:            1,
		Pem:              "pem",
	}

	r, err := provider.RegisterRootCertificate(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer provider.RemoveRootCertificate(ctx, r.ID)

	assert.Equal(t, o.OwnerID, r.OwnerID)
	assert.Equal(t, o.Skid, r.Skid)
	assert.Equal(t, o.ThumbprintSha256, r.ThumbprintSha256)
	assert.Equal(t, o.Trust, r.Trust)
	assert.Equal(t, o.Pem, r.Pem)
	assert.Equal(t, o.NotBefore.Unix(), r.NotBefore.Unix())
	assert.Equal(t, o.NotAfter.Unix(), r.NotAfter.Unix())

	r2, err := provider.RegisterRootCertificate(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, *r, *r2)

	list, err := provider.GetRootCertificates(ctx, ownerID)
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
