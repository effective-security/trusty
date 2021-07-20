package pgsql_test

import (
	"fmt"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoginUser(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	u1 := &model.User{
		Provider:   v1.ProviderGithub,
		Name:       name,
		Login:      login,
		Email:      email,
		ExternalID: fmt.Sprintf("%d", id+13),
	}
	u2 := &model.User{
		Provider:   v1.ProviderGoogle,
		Name:       name,
		Login:      login,
		Email:      email,
		ExternalID: fmt.Sprintf("%d", id+17),
	}

	user, err := provider.LoginUser(ctx, u1)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, v1.ProviderGithub, user.Provider)
	assert.Equal(t, 1, user.LoginCount)

	guser, err := provider.LoginUser(ctx, u2)
	require.NoError(t, err)
	assert.NotNil(t, guser)
	assert.Equal(t, name, guser.Name)
	assert.Equal(t, email, guser.Email)
	assert.Equal(t, login, guser.Login)
	assert.Equal(t, v1.ProviderGoogle, guser.Provider)
	assert.Equal(t, 1, guser.LoginCount)

	require.NotEqual(t, user.ID, guser.ID)

	user2, err := provider.LoginUser(ctx, u1)
	require.NoError(t, err)
	assert.NotNil(t, user2)
	assert.Equal(t, name, user2.Name)
	assert.Equal(t, login, user2.Login)
	assert.Equal(t, email, user2.Email)
	assert.Equal(t, 2, user2.LoginCount)

	assert.Equal(t, user.ID, user2.ID)

	user3, err := provider.GetUser(ctx, user2.ID)
	require.NoError(t, err)
	require.NotNil(t, user3)
	assert.Equal(t, name, user3.Name)
	assert.Equal(t, login, user3.Login)
	assert.Equal(t, email, user3.Email)
	assert.Equal(t, 2, user3.LoginCount)
	/*
		list, err := provider.ListUsers(ctx, "", 100)
		require.NoError(t, err)
		require.NotEmpty(t, list)
	*/
}
