package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApprovalTokens(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	name := fmt.Sprintf("user-%d", id)
	login := fmt.Sprintf("test%d", id)
	email := login + "@trusty.com"

	o := &model.Organization{
		ExternalID: fmt.Sprintf("%d", id+2),
		Provider:   v1.ProviderGithub,
		Name:       name,
		Login:      login,
		Email:      email,
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

	token := fmt.Sprintf("t-%d", id)

	m := &model.ApprovalToken{
		OrgID:         org.ID,
		RequestorID:   id + 1,
		ApproverEmail: token + "@ekspand.com",
		Token:         token[:16],
		Code:          fmt.Sprintf("%d", id)[:6],
		Used:          false,
		CreatedAt:     time.Now().UTC(),
	}

	m1, err := provider.CreateApprovalToken(ctx, m)
	require.NoError(t, err)
	require.NotNil(t, m1)
	defer provider.DeleteApprovalToken(ctx, m1.ID)

	assert.Equal(t, m.OrgID, m1.OrgID)
	assert.Equal(t, m.RequestorID, m1.RequestorID)
	assert.Equal(t, m.ApproverEmail, m1.ApproverEmail)
	assert.Equal(t, m.Token, m1.Token)
	assert.Equal(t, m.Code, m1.Code)
	assert.Equal(t, m.Used, m1.Used)
	assert.Equal(t, m.CreatedAt.Unix(), m1.CreatedAt.Unix())
	assert.Equal(t, m.ExpiresAt.Unix(), m1.ExpiresAt.Unix())
	assert.Equal(t, m.UsedAt.Unix(), m1.UsedAt.Unix())

	org2, err := provider.GetOrgFromApprovalToken(ctx, m1.Token)
	require.NoError(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, *org, *org2)

	_, err = provider.CreateApprovalToken(ctx, m)
	require.Error(t, err)
	assert.Equal(t, "pq: duplicate key value violates unique constraint \"orgtokens_token_code\"", err.Error())

	_, err = provider.UseApprovalToken(ctx, "notfound", "123456")
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	m2, err := provider.UseApprovalToken(ctx, m.Token, m.Code)
	require.NoError(t, err)
	assert.True(t, m2.Used)
	assert.Equal(t, m.OrgID, m2.OrgID)
	assert.Equal(t, m.RequestorID, m2.RequestorID)
	assert.Equal(t, m.ApproverEmail, m2.ApproverEmail)
	assert.Equal(t, m.Token, m2.Token)
	assert.Equal(t, m.Code, m2.Code)

	_, err = provider.UseApprovalToken(ctx, m.Token, m.Code)
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	list, err := provider.GetOrgApprovalTokens(ctx, m.OrgID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}
