package pgsql_test

import (
	"fmt"
	"testing"

	"github.com/go-phorce/trusty/internal/db/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoginUser(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	u := &model.User{
		Name:  name,
		Login: login,
		Email: email,
	}

	user, err := provider.LoginUser(ctx, u)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, 1, user.LoginCount)

	user2, err := provider.LoginUser(ctx, u)
	require.NoError(t, err)
	assert.NotNil(t, user2)
	assert.Equal(t, name, user2.Name)
	assert.Equal(t, login, user2.Login)
	assert.Equal(t, email, user2.Email)
	assert.Equal(t, 2, user2.LoginCount)

	assert.Equal(t, user.ID, user2.ID)
	/*
		list, err := provider.ListUsers(ctx, "", 100)
		require.NoError(t, err)
		require.NotEmpty(t, list)
	*/
}
