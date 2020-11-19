package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateOrg(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	uid := int64(id)
	o := &model.Organization{
		ExternalID: model.NullInt64(&uid),
		Provider:   "github",
		Name:       name,
		Login:      login,
		Email:      email,
		Company:    "ekspand",
		Type:       "public",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	org, err := provider.UpdateOrg(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, name, org.Name)
	assert.Equal(t, email, org.Email)
	assert.Equal(t, login, org.Login)

	org2, err := provider.UpdateOrg(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, *org, *org2)

	org3, err := provider.GetOrg(ctx, org2.ID)
	require.NoError(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, *org, *org3)
}

func TestRepositoryOrg(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	uid := int64(id)
	o := &model.Repository{
		OrgID:      uid,
		ExternalID: model.NullInt64(&uid),
		Provider:   "github",
		Name:       name,
		Email:      email,
		Company:    "ekspand",
		Type:       "public",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	repo, err := provider.UpdateRepo(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, name, repo.Name)
	assert.Equal(t, email, repo.Email)

	repo2, err := provider.UpdateRepo(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, repo2)
	assert.Equal(t, *repo, *repo2)

	repo3, err := provider.GetRepo(ctx, repo2.ID)
	require.NoError(t, err)
	require.NotNil(t, repo2)
	assert.Equal(t, *repo, *repo3)
}
