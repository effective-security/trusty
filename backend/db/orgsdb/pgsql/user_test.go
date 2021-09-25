package pgsql_test

import (
	"fmt"
	"testing"

	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginUser(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	u1 := &model.User{
		Provider:    v1.ProviderGithub,
		Name:        name,
		Login:       login,
		Email:       email,
		Company:     "c1",
		AvatarURL:   "https://av.com/123",
		AccessToken: "12334",
		ExternalID:  fmt.Sprintf("%d", id+13),
	}
	u2 := &model.User{
		Provider:    v1.ProviderGoogle,
		Name:        name,
		Login:       login,
		Email:       email,
		Company:     "c1",
		AvatarURL:   "https://av.com/123",
		AccessToken: "12335",
		ExternalID:  fmt.Sprintf("%d", id+17),
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

	user4, err := provider.CreateUser(ctx, user2.Provider, user2.Email)
	require.NoError(t, err)
	assert.NotNil(t, user4)
	assert.Equal(t, user2.Name, user4.Name)
	assert.Equal(t, user2.ExternalID, user4.ExternalID)
	assert.Equal(t, user2.Company, user4.Company)
	assert.Equal(t, user2.AvatarURL, user4.AvatarURL)

	/*
		list, err := provider.ListUsers(ctx, "", 100)
		require.NoError(t, err)
		require.NotEmpty(t, list)
	*/

	c, err := provider.GetUsersCount(ctx)
	require.NoError(t, err)
	assert.Greater(t, c, uint64(0))
}

func TestCreateUser(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	u1 := &model.User{
		Provider:    v1.ProviderGoogle,
		Name:        name,
		Login:       login,
		Email:       email,
		Company:     "c1",
		AvatarURL:   "https://av.com/123",
		AccessToken: "12334",
		ExternalID:  fmt.Sprintf("%d", id+13),
	}

	user1, err := provider.CreateUser(ctx, u1.Provider, u1.Email)
	require.NoError(t, err)
	assert.NotNil(t, user1)
	assert.Equal(t, u1.Email, user1.Email)
	assert.Equal(t, u1.Email, user1.Name)
	assert.Equal(t, u1.Email, user1.Login)
	assert.Equal(t, db.IDString(user1.ID), user1.ExternalID)
	assert.Empty(t, user1.Company)
	assert.Empty(t, user1.AvatarURL)
	assert.Empty(t, user1.AccessToken)
	assert.Equal(t, 0, user1.LoginCount)

	user2, err := provider.LoginUser(ctx, u1)
	require.NoError(t, err)
	assert.NotNil(t, user2)
	assert.Equal(t, name, user2.Name)
	assert.Equal(t, email, user2.Email)
	assert.Equal(t, login, user2.Login)
	assert.Equal(t, u1.ExternalID, user2.ExternalID)
	assert.Equal(t, u1.Company, user2.Company)
	assert.Equal(t, u1.AvatarURL, user2.AvatarURL)
	assert.Equal(t, u1.Provider, user2.Provider)
	assert.Equal(t, 1, user2.LoginCount)
}
