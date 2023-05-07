package pgsql_test

import (
	"testing"
	"time"

	"github.com/effective-security/porto/x/guid"
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRootCertificate(t *testing.T) {
	rc := &model.RootCertificate{
		SKID:             guid.MustCreate(),
		Subject:          "subj",
		NotBefore:        xdb.FromNow(-time.Hour),
		NotAfter:         xdb.FromNow(time.Hour),
		ThumbprintSha256: certutil.RandomString(64),
		Trust:            1,
		Pem:              "pem",
	}

	r, err := provider.RegisterRootCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() {
		_ = provider.RemoveRootCertificate(ctx, r.ID)
	}()

	assert.Equal(t, rc.SKID, r.SKID)
	assert.Equal(t, rc.Subject, r.Subject)
	assert.Equal(t, rc.ThumbprintSha256, r.ThumbprintSha256)
	assert.Equal(t, rc.Trust, r.Trust)
	assert.Equal(t, rc.Pem, r.Pem)
	assert.Equal(t, rc.NotBefore, r.NotBefore)
	assert.Equal(t, rc.NotAfter, r.NotAfter)

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

func TestGetCertificatesBySKID(t *testing.T) {
	_, err := provider.GetCertificatesBySKID(ctx, "notfound")
	assert.EqualError(t, err, "sql: no rows in result set")
	assert.True(t, xdb.IsNotFoundError(err))
}

func TestRegisterCertificate(t *testing.T) {
	orgID := provider.NextID()

	rc := &model.Certificate{
		OrgID:            orgID,
		SKID:             guid.MustCreate(),
		IKID:             guid.MustCreate(),
		SerialNumber:     certutil.RandomString(10),
		Subject:          "subj",
		Issuer:           "iss",
		NotBefore:        xdb.FromNow(-time.Hour),
		NotAfter:         xdb.FromNow(time.Hour),
		ThumbprintSha256: certutil.RandomString(64),
		Pem:              "pem",
		IssuersPem:       "ipem",
		Profile:          "client",
		Label:            "label",
	}

	r, err := provider.RegisterCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() {
		_ = provider.RemoveCertificate(ctx, r.ID)
	}()

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
	assert.Equal(t, rc.Label, r.Label)
	assert.Equal(t, rc.Locations, r.Locations)
	assert.Equal(t, rc.NotBefore, r.NotBefore)
	assert.Equal(t, rc.NotAfter, r.NotAfter)
	assert.Empty(t, rc.Locations)
	assert.Empty(t, rc.Metadata)

	r.Locations = []string{"1", "2", "3"}
	r.Metadata = map[string]string{"requester": "denis"}

	r2, err := provider.RegisterCertificate(ctx, r)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, *r, *r2)
	assert.Len(t, r2.Locations, 3)
	assert.Len(t, r2.Metadata, 1)

	rcx := *r2
	rcx.ThumbprintSha256 = certutil.RandomString(64)
	// the same IKID, Serial
	_, err = provider.RegisterCertificate(ctx, &rcx)
	assert.EqualError(t, err, "pq: duplicate key value violates unique constraint \"idx_certificates_ikid_serial\"")
	// another SN
	rcx.SerialNumber = certutil.RandomString(10)
	_, err = provider.RegisterCertificate(ctx, &rcx)
	assert.NoError(t, err)

	list, err := provider.ListOrgCertificates(ctx, orgID, 10, 0)
	require.NoError(t, err)
	r3 := list.Find(r.ID)
	require.NotNil(t, r3)
	assert.NotEqual(t, *r, *r3)
	r3c := *r3
	r3c.IssuersPem = r.IssuersPem
	assert.Equal(t, *r, r3c)

	list2, err := provider.ListCertificates(ctx, r3.IKID, 100, 0)
	require.NoError(t, err)
	c := list2.Find(r.ID)
	require.NotNil(t, c)
	assert.NotEqual(t, *r, *c)

	last := list2[len(list2)-1]
	list2, err = provider.ListCertificates(ctx, r3.IKID, 100, last.ID)
	require.NoError(t, err)
	assert.Empty(t, list2)

	r4, err := provider.GetCertificate(ctx, r2.ID)
	require.NoError(t, err)
	require.NotNil(t, r4)
	assert.Equal(t, *r, *r4)

	r4l, err := provider.GetCertificatesBySKID(ctx, r2.SKID)
	require.NoError(t, err)
	assert.Len(t, r4l, 2)
	//assert.Equal(t, *r, *r4l[0])

	r4, err = provider.GetCertificateByIKIDAndSerial(ctx, r2.IKID, r2.SerialNumber)
	require.NoError(t, err)
	require.NotNil(t, r4)
	assert.Equal(t, *r, *r4)

	r5, err := provider.UpdateCertificateLabel(ctx, r4.ID, r4.Label+"1")
	require.NoError(t, err)
	assert.Equal(t, r4.Label+"1", r5.Label)

	cc, err := provider.GetTableRowsCount(ctx, cadb.TableNameForCertificates)
	require.NoError(t, err)
	assert.Greater(t, cc, uint64(0))

	cr, err := provider.GetTableRowsCount(ctx, cadb.TableNameForRevoked)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cr, uint64(0))

	revoked, err := provider.RevokeCertificate(ctx, r4, time.Now(), 0)
	require.NoError(t, err)
	revoked.Certificate.Locations = r.Locations
	assert.Equal(t, revoked.Certificate, *r4)

	revoked2, err := provider.GetRevokedCertificateByIKIDAndSerial(ctx, r4.IKID, r4.SerialNumber)
	require.NoError(t, err)
	revoked2.Certificate.Locations = r.Locations
	assert.Equal(t, *revoked, *revoked2)

	cc2, err := provider.GetTableRowsCount(ctx, cadb.TableNameForCertificates)
	require.NoError(t, err)
	assert.Greater(t, cc, cc2)

	cr2, err := provider.GetTableRowsCount(ctx, cadb.TableNameForRevoked)
	require.NoError(t, err)
	assert.Greater(t, cr2, cr)

	_, err = provider.GetCertificate(ctx, r2.ID)
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	err = provider.RemoveRevokedCertificate(ctx, revoked.Certificate.ID)
	require.NoError(t, err)
}

func TestRegisterCertificateUniqueIdx(t *testing.T) {
	orgID := provider.NextID()
	rc := &model.Certificate{
		OrgID:            orgID,
		SKID:             guid.MustCreate(),
		IKID:             guid.MustCreate(),
		SerialNumber:     certutil.RandomString(10),
		Subject:          "subj",
		Issuer:           "iss",
		NotBefore:        xdb.FromNow(-time.Hour),
		NotAfter:         xdb.FromNow(time.Hour),
		ThumbprintSha256: certutil.RandomString(64),
		Pem:              "pem",
		IssuersPem:       "ipem",
		Profile:          "client",
		Label:            "label",
		Locations:        []string{"l1", "l2"},
		Metadata:         map[string]string{"requester": "test"},
	}

	r, err := provider.RegisterCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() {
		_ = provider.RemoveCertificate(ctx, r.ID)
	}()

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
	assert.Equal(t, rc.Label, r.Label)
	assert.Equal(t, rc.Locations, r.Locations)
	assert.Equal(t, rc.Locations, r.Locations)
	assert.Equal(t, rc.Metadata, r.Metadata)
	assert.Equal(t, rc.NotBefore, r.NotBefore)
	assert.Equal(t, rc.NotAfter, r.NotAfter)

	r2, err := provider.RegisterCertificate(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r2)
	assert.EqualValues(t, *r, *r2)

	// change sha2
	rc.ThumbprintSha256 = certutil.RandomString(64)
	_, err = provider.RegisterCertificate(ctx, rc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pq: duplicate key value violates unique constraint \"idx_certificates_ikid_serial\"")

	// change skid
	rc.SKID = guid.MustCreate()
	_, err = provider.RegisterCertificate(ctx, rc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pq: duplicate key value violates unique constraint \"idx_certificates_ikid_serial\"")
}

func TestRegisterRevokedCertificate(t *testing.T) {
	orgID := provider.NextID()

	mrc := &model.RevokedCertificate{
		Certificate: model.Certificate{
			OrgID:            orgID,
			SKID:             guid.MustCreate(),
			IKID:             guid.MustCreate(),
			SerialNumber:     certutil.RandomString(10),
			Subject:          "subj",
			Issuer:           "iss",
			NotBefore:        xdb.FromNow(-time.Hour),
			NotAfter:         xdb.FromNow(time.Hour),
			ThumbprintSha256: certutil.RandomString(64),
			Pem:              "pem",
			IssuersPem:       "ipem",
			Profile:          "client",
			Label:            "label",
			Locations:        []string{"l1", "l2"},
			Metadata:         map[string]string{"host": "local"},
		},
		RevokedAt: xdb.FromNow(0),
		Reason:    1,
	}

	mr, err := provider.RegisterRevokedCertificate(ctx, mrc)
	require.NoError(t, err)
	require.NotNil(t, mr)
	defer func() {
		_ = provider.RemoveRevokedCertificate(ctx, mr.Certificate.ID)
	}()

	rc := &mrc.Certificate
	r := &mr.Certificate
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
	assert.Equal(t, rc.Label, r.Label)
	assert.Equal(t, rc.Locations, r.Locations)
	assert.Equal(t, rc.Metadata, r.Metadata)
	assert.Equal(t, rc.NotBefore, r.NotBefore)
	assert.Equal(t, rc.NotAfter, r.NotAfter)

	list, err := provider.ListOrgRevokedCertificates(ctx, orgID, 100, 0)
	require.NoError(t, err)
	r3 := list.Find(r.ID)
	require.NotNil(t, r3)
	assert.NotEqual(t, *mr, *r3)
	r3c := *r3
	r3c.Certificate.Pem = r.Pem
	r3c.Certificate.IssuersPem = r.IssuersPem
	assert.Equal(t, *mr, r3c)

	list, err = provider.ListRevokedCertificates(ctx, r.IKID, 0, 0)
	require.NoError(t, err)
	r4 := list.Find(r.ID)
	require.NotNil(t, r4)
	assert.NotEqual(t, *mr, *r4)
	// Pems are not returned by List
	mr.Certificate.Pem = ""
	mr.Certificate.IssuersPem = ""
	assert.Equal(t, *mr, *r4)
}

func TestRegisterCrl(t *testing.T) {
	rc := &model.Crl{
		IKID:       guid.MustCreate(),
		Issuer:     "iss",
		ThisUpdate: xdb.FromNow(-time.Hour),
		NextUpdate: xdb.FromNow(time.Hour),
		Pem:        "pem",
	}

	r, err := provider.RegisterCrl(ctx, rc)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() {
		_ = provider.RemoveCrl(ctx, r.ID)
	}()

	assert.Equal(t, rc.IKID, r.IKID)
	assert.Equal(t, rc.Issuer, r.Issuer)
	assert.Equal(t, rc.Pem, r.Pem)
	assert.Equal(t, rc.ThisUpdate, r.ThisUpdate)
	assert.Equal(t, rc.NextUpdate, r.NextUpdate)

	r2, err := provider.GetCrl(ctx, r.IKID)
	require.NoError(t, err)
	assert.Equal(t, rc.IKID, r2.IKID)
	assert.Equal(t, rc.Issuer, r2.Issuer)
	assert.Equal(t, rc.Pem, r2.Pem)
	assert.Equal(t, rc.ThisUpdate, r2.ThisUpdate)
	assert.Equal(t, rc.NextUpdate, r2.NextUpdate)
}

func TestListCertificate(t *testing.T) {
	count := 20
	orgID := uint64(1000)
	ikid := guid.MustCreate()

	for i := 0; i < count; i++ {
		rc := &model.Certificate{
			OrgID:            orgID,
			SKID:             guid.MustCreate(),
			IKID:             ikid,
			SerialNumber:     certutil.RandomString(10),
			Subject:          "subj",
			Issuer:           "iss",
			NotBefore:        xdb.FromNow(-time.Hour),
			NotAfter:         xdb.FromNow(time.Hour),
			ThumbprintSha256: certutil.RandomString(64),
			Pem:              "pem",
			IssuersPem:       "ipem",
			Profile:          "client",
			Label:            "label",
			Locations:        []string{"l1", "l2"},
		}

		r, err := provider.RegisterCertificate(ctx, rc)
		require.NoError(t, err)
		require.NotNil(t, r)
		defer func() {
			_ = provider.RemoveCertificate(ctx, r.ID)
		}()
	}

	list, err := provider.ListCertificates(ctx, ikid, 1000, 0)
	require.NoError(t, err)
	require.Len(t, list, count)

	first := list[0].ID
	last := list[count-1].ID

	list2, err := provider.ListCertificates(ctx, ikid, 1000, first)
	require.NoError(t, err)
	require.Len(t, list2, count-1)
	assert.Nil(t, list2.Find(first))

	list3, err := provider.ListCertificates(ctx, ikid, 1000, last)
	require.NoError(t, err)
	require.Len(t, list3, 0)

	last = uint64(0)
	bulk := make([]*model.Certificate, 0, count)
	for {
		list3, err = provider.ListCertificates(ctx, ikid, 5, last)
		require.NoError(t, err)
		if len(list3) == 0 {
			break
		}

		bulk = append(bulk, list3...)
		last = list3[len(list3)-1].ID
	}
	require.Len(t, bulk, count)
}
