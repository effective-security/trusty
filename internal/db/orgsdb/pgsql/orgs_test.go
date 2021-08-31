package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateOrg(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	o := &model.Organization{
		ExternalID:     fmt.Sprintf("%d", id+2),
		RegistrationID: fmt.Sprintf("%d", id+3),
		Provider:       v1.ProviderGithub,
		Name:           name,
		Login:          login,
		Email:          email,
		//BillingEmail: email,
		Company:       "ekspand",
		Location:      "Kirkland, WA",
		Type:          "Organization",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Street:        "addr",
		City:          "city",
		PostalCode:    "98034",
		Region:        "WA",
		Country:       "US",
		Phone:         "4251232323",
		ApproverName:  "approver",
		ApproverEmail: "denis@ekspand.com",
		Status:        "pending",
		ExpiresAt:     time.Now().Add(time.Hour * 8770).UTC(),
	}

	org, err := provider.UpdateOrg(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, org)
	defer provider.RemoveOrg(ctx, org.ID)

	assert.Equal(t, name, org.Name)
	assert.Equal(t, email, org.Email)
	assert.Empty(t, org.BillingEmail)
	assert.Equal(t, login, org.Login)
	assert.Equal(t, o.Location, org.Location)
	assert.Equal(t, o.Type, org.Type)
	assert.Equal(t, o.Company, org.Company)
	assert.Equal(t, o.Street, org.Street)
	assert.Equal(t, o.City, org.City)
	assert.Equal(t, o.PostalCode, org.PostalCode)
	assert.Equal(t, o.Region, org.Region)
	assert.Equal(t, o.Country, org.Country)
	assert.Equal(t, o.Phone, org.Phone)
	assert.Equal(t, o.ApproverEmail, org.ApproverEmail)
	assert.Equal(t, o.ApproverName, org.ApproverName)
	assert.Equal(t, o.Status, org.Status)
	assert.Equal(t, o.ExternalID, org.ExternalID)
	assert.Equal(t, o.RegistrationID, org.RegistrationID)

	org.Company = "Ekspand"
	org.BillingEmail = email
	org2, err := provider.UpdateOrg(ctx, org)
	require.NoError(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, *org, *org2)

	org3, err := provider.GetOrg(ctx, org2.ID)
	require.NoError(t, err)
	require.NotNil(t, org3)
	assert.Equal(t, *org, *org3)

	org.Status = "approved"
	org4, err := provider.UpdateOrgStatus(ctx, org)
	require.NoError(t, err)
	require.NotNil(t, org4)
	assert.Equal(t, *org, *org4)

	org5, err := provider.GetOrgByExternalID(ctx, o.Provider, o.ExternalID)
	require.NoError(t, err)
	require.NotNil(t, org5)
	assert.Equal(t, *org, *org5)

	c, err := provider.GetOrgsCount(ctx)
	require.NoError(t, err)
	assert.Greater(t, c, uint64(0))
}

func TestRepositoryOrg(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	o := &model.Repository{
		OrgID:      id,
		ExternalID: id,
		Provider:   v1.ProviderGithub,
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

func Test_Membership(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	login1 := fmt.Sprintf("user1%d", id)
	login2 := fmt.Sprintf("user2%d", id)
	email1 := fmt.Sprintf("test1%d@ekspand.com", id)
	email2 := fmt.Sprintf("test2%d@ekspand.com", id)
	name := fmt.Sprintf("org-%d", id)

	o := &model.Organization{
		ExternalID:     fmt.Sprintf("%d", id),
		RegistrationID: fmt.Sprintf("%d", id+1),
		Provider:       v1.ProviderGithub,
		Name:           name,
		Login:          email1,
		Email:          email1,
		//BillingEmail: email,
		Company:   "ekspand",
		Location:  "Kirkland, WA",
		Type:      "Organization",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	org, err := provider.UpdateOrg(ctx, o)
	require.NoError(t, err)
	require.NotNil(t, org)
	defer provider.RemoveOrg(ctx, org.ID)

	oldMembers, err := provider.RemoveOrgMembers(ctx, org.ID, true)
	require.NoError(t, err)
	assert.Empty(t, oldMembers)

	user1, err := provider.LoginUser(ctx, &model.User{
		Login:      login1,
		Email:      email1,
		Name:       email1,
		ExternalID: fmt.Sprintf("%d", id+1),
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)
	assert.NotNil(t, user1)
	assert.Equal(t, email1, user1.Name)
	assert.Equal(t, email1, user1.Email)
	assert.Equal(t, 1, user1.LoginCount)

	user2, err := provider.LoginUser(ctx, &model.User{
		Login:      login2,
		Email:      email2,
		Name:       email2,
		ExternalID: fmt.Sprintf("%d", id+2),
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)
	assert.NotNil(t, user2)
	assert.Equal(t, email2, user2.Name)
	assert.Equal(t, email2, user2.Email)
	assert.Equal(t, 1, user2.LoginCount)

	ms, err := provider.AddOrgMember(ctx, org.ID, user1.ID, "admin", v1.ProviderGithub)
	require.NoError(t, err)
	assert.Equal(t, "admin", ms.Role)
	assert.Equal(t, user1.ID, ms.UserID)
	assert.Equal(t, org.ID, ms.OrgID)

	ms, err = provider.AddOrgMember(ctx, org.ID, user2.ID, "user", v1.ProviderGithub)
	require.NoError(t, err)
	assert.Equal(t, "user", ms.Role)
	assert.Equal(t, user2.ID, ms.UserID)
	assert.Equal(t, org.ID, ms.OrgID)

	list, err := provider.GetOrgMembers(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list))
	m := findOrgMember(list, user1.ID)
	assert.NotNil(t, m)
	assert.Equal(t, user1.Email, m.Email)
	assert.Equal(t, user1.Name, m.Name)
	assert.Equal(t, org.Name, m.OrgName)

	list2, err := provider.GetUserMemberships(ctx, user1.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(list2))
	assert.Equal(t, org.Name, list2[0].OrgName)

	orgs, err := provider.GetUserOrgs(ctx, user1.ID)
	require.NoError(t, err)
	assert.Len(t, orgs, 1)

	removed, err := provider.RemoveOrgMember(ctx, org.ID, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, "user", removed.Role)
	assert.Equal(t, user2.ID, removed.UserID)
	assert.Equal(t, org.ID, removed.OrgID)

	ms, err = provider.AddOrgMember(ctx, org.ID, user2.ID, "user", v1.ProviderGithub)
	require.NoError(t, err)
	assert.Equal(t, "user", ms.Role)
	assert.Equal(t, user2.ID, ms.UserID)
	assert.Equal(t, org.ID, ms.OrgID)

	oldMembers, err = provider.RemoveOrgMembers(ctx, org.ID, true)
	require.NoError(t, err)
	assert.Equal(t, len(list), len(oldMembers))
}

func findOrgMember(list []*model.OrgMemberInfo, userID uint64) *model.OrgMemberInfo {
	for _, m := range list {
		if m.UserID == userID {
			return m
		}
	}
	return nil
}

func findUser(list []*model.User, userID uint64) *model.User {
	for _, m := range list {
		if m.ID == userID {
			return m
		}
	}
	return nil
}

func findOrg(list []*model.Organization, orgID uint64) *model.Organization {
	for _, m := range list {
		if m.ID == orgID {
			return m
		}
	}
	return nil
}
