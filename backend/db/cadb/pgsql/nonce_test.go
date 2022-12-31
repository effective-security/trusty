package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonces(t *testing.T) {
	id := provider.NextID()

	token := fmt.Sprintf("t-%d", id)

	m := &model.Nonce{
		Nonce:     token[:16],
		Used:      false,
		CreatedAt: time.Now().UTC(),
	}

	m1, err := provider.CreateNonce(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m1)
	_ = provider.DeleteNonce(ctx, m1.ID)

	assert.Equal(t, m.Nonce, m1.Nonce)
	assert.Equal(t, m.Used, m1.Used)
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.UsedAt.Unix(), m1.UsedAt.Unix())

	_, err = provider.CreateNonce(ctx, m)
	require.Error(t, err)
	assert.Equal(t, "pq: duplicate key value violates unique constraint \"nonces_nonce\"", err.Error())

	_, err = provider.UseNonce(ctx, "notfound")
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	m2, err := provider.UseNonce(ctx, m.Nonce)
	require.NoError(t, err)
	assert.True(t, m2.Used)
	assert.Equal(t, m.Nonce, m2.Nonce)

	_, err = provider.UseNonce(ctx, m.Nonce)
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())
}
