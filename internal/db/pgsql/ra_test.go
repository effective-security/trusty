package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRootCertificate(t *testing.T) {
	rc := &model.RootCertificate{
		SKID:             guid.MustCreate(),
		Subject:          "subj",
		NotBefore:        time.Now().Add(-time.Hour).UTC(),
		NotAfter:         time.Now().Add(time.Hour).UTC(),
		ThumbprintSha256: certutil.RandomString(64),
		Trust:            1,
		Pem:              "pem",
	}

	r, err := provider.RegisterRootCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer provider.RemoveRootCertificate(ctx, r.ID)

	assert.Equal(t, rc.SKID, r.SKID)
	assert.Equal(t, rc.Subject, r.Subject)
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
	r3 := list.Find(r.ID)
	require.NotNil(t, r3)
	assert.Equal(t, *r, *r3)
}

func TestRegisterCertificate(t *testing.T) {
	id, err := provider.NextID()
	require.NoError(t, err)

	login1 := fmt.Sprintf("user1%d", id)
	email1 := fmt.Sprintf("test1%d@ekspand.com", id)
	name := fmt.Sprintf("org-%d", id)

	uid := int64(id)
	or := &model.Organization{
		ExternalID: uid,
		Provider:   v1.ProviderGithub,
		Name:       name,
		Login:      email1,
		Email:      email1,
		//BillingEmail: email,
		Company:   "ekspand",
		Location:  "Kirkland, WA",
		Type:      "Organization",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	org, err := provider.UpdateOrg(ctx, or)
	require.NoError(t, err)
	require.NotNil(t, org)
	defer provider.RemoveOrg(ctx, org.ID)

	user1, err := provider.LoginUser(ctx, &model.User{Login: login1, Email: email1, Name: email1})
	require.NoError(t, err)
	assert.NotNil(t, user1)

	_, err = provider.AddOrgMember(ctx, org.ID, user1.ID, "admin", v1.ProviderGithub)
	require.NoError(t, err)

	rc := &model.Certificate{
		OrgID:            org.ID,
		SKID:             guid.MustCreate(),
		IKID:             guid.MustCreate(),
		SerialNumber:     certutil.RandomString(10),
		Subject:          "subj",
		Issuer:           "iss",
		NotBefore:        time.Now().Add(-time.Hour).UTC(),
		NotAfter:         time.Now().Add(time.Hour).UTC(),
		ThumbprintSha256: certutil.RandomString(64),
		Pem:              "pem",
		IssuersPem:       "ipem",
		Profile:          "client",
	}

	r, err := provider.RegisterCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer provider.RemoveCertificate(ctx, r.ID)

	assert.Equal(t, rc.OrgID, r.OrgID)
	assert.Equal(t, rc.SKID, r.SKID)
	assert.Equal(t, rc.IKID, r.IKID)
	assert.Equal(t, rc.SerialNumber, r.SerialNumber)
	assert.Equal(t, rc.Subject, r.Subject)
	assert.Equal(t, rc.Issuer, r.Issuer)
	assert.Equal(t, rc.ThumbprintSha256, r.ThumbprintSha256)
	assert.Equal(t, rc.Pem, r.Pem)
	assert.Equal(t, rc.IssuersPem, r.IssuersPem)
	assert.Equal(t, rc.Profile, r.Profile)
	assert.Equal(t, rc.NotBefore.Unix(), r.NotBefore.Unix())
	assert.Equal(t, rc.NotAfter.Unix(), r.NotAfter.Unix())

	r2, err := provider.RegisterCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, *r, *r2)

	list, err := provider.GetCertificatesForOrg(ctx, org.ID)
	require.NoError(t, err)
	r3 := list.Find(r.ID)
	require.NotNil(t, r3)
	assert.Equal(t, *r, *r3)

	list, err = provider.GetCertificatesForUser(ctx, user1.ID)
	require.NoError(t, err)
	r4 := list.Find(r.ID)
	require.NotNil(t, r4)
	assert.Equal(t, *r, *r4)
}
