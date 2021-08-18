package model_test

import (
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrl(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	m := &model.Crl{
		ID:         123,
		IKID:       "ikid",
		ThisUpdate: nb.UTC(),
		NextUpdate: na.UTC(),
		Issuer:     "issuer",
		Pem:        "pem",
	}

	dto := m.ToDTO()
	assert.Equal(t, uint64(123), dto.Id)
	assert.Equal(t, m.IKID, dto.Ikid)
	assert.Equal(t, m.ThisUpdate, dto.ThisUpdate.AsTime().UTC())
	assert.Equal(t, m.NextUpdate, dto.NextUpdate.AsTime().UTC())
	assert.Equal(t, m.Issuer, dto.Issuer)
	assert.Equal(t, m.Pem, dto.Pem)
}

func TestCertificate(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	m := &model.Certificate{
		ID:               123,
		OrgID:            234,
		SKID:             "skid",
		IKID:             "ikid",
		SerialNumber:     "1231234123121232",
		NotBefore:        nb.UTC(),
		NotAfter:         na.UTC(),
		Subject:          "subject",
		Issuer:           "issuer",
		ThumbprintSha256: "sha256",
		Profile:          "profile",
		Pem:              "pem",
		IssuersPem:       "issuers_pem",
		Locations:        []string{"1"},
	}
	dto := m.ToPB()
	assert.Equal(t, uint64(123), dto.Id)
	assert.Equal(t, uint64(234), dto.OrgId)
	assert.Equal(t, m.SKID, dto.Skid)
	assert.Equal(t, m.IKID, dto.Ikid)
	assert.Equal(t, m.SerialNumber, dto.SerialNumber)
	assert.Equal(t, m.NotBefore, dto.NotBefore.AsTime().UTC())
	assert.Equal(t, m.NotAfter, dto.NotAfter.AsTime().UTC())
	assert.Equal(t, m.Subject, dto.Subject)
	assert.Equal(t, m.Issuer, dto.Issuer)
	assert.Equal(t, m.ThumbprintSha256, dto.Sha256)
	assert.Equal(t, m.Profile, dto.Profile)
	assert.Equal(t, m.Pem, dto.Pem)
	assert.Equal(t, m.IssuersPem, dto.IssuersPem)
	assert.Equal(t, m.Locations, dto.Locations)

	fn := m.FileName()
	assert.Contains(t, fn, "/")

	m2 := model.CertificateFromPB(dto)
	assert.Equal(t, *m, *m2)

	l := model.Certificates{m}
	assert.Len(t, l.ToDTO(), 1)

	m3 := l.Find(m.ID)
	assert.Equal(t, *m, *m3)

	crt, err := certutil.ParseFromPEM([]byte(testCrt))
	require.NoError(t, err)
	m4 := model.NewCertificate(crt, 123, "ca", testCrt, testCrt, nil)
	assert.Equal(t, uint64(0), m4.ID)
}

func TestRevokedCertificate(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	r := &model.RevokedCertificate{
		Certificate: model.Certificate{
			ID:               123,
			OrgID:            234,
			SKID:             "skid",
			IKID:             "ikid",
			SerialNumber:     "serial_number",
			NotBefore:        nb.UTC(),
			NotAfter:         na.UTC(),
			Subject:          "subject",
			Issuer:           "issuer",
			ThumbprintSha256: "sha256",
			Profile:          "profile",
			Pem:              "pem",
			IssuersPem:       "issuers_pem",
		},
		Reason: 1,
	}
	dto := r.ToDTO()
	m := r.Certificate
	pbc := dto.Certificate
	assert.Equal(t, uint64(123), pbc.Id)
	assert.Equal(t, uint64(234), pbc.OrgId)
	assert.Equal(t, m.SKID, pbc.Skid)
	assert.Equal(t, m.IKID, pbc.Ikid)
	assert.Equal(t, m.SerialNumber, pbc.SerialNumber)
	assert.Equal(t, m.NotBefore, pbc.NotBefore.AsTime().UTC())
	assert.Equal(t, m.NotAfter, pbc.NotAfter.AsTime().UTC())
	assert.Equal(t, m.Subject, pbc.Subject)
	assert.Equal(t, m.Issuer, pbc.Issuer)
	assert.Equal(t, m.ThumbprintSha256, pbc.Sha256)
	assert.Equal(t, m.Profile, pbc.Profile)
	assert.Equal(t, m.Pem, pbc.Pem)
	assert.Equal(t, m.IssuersPem, pbc.IssuersPem)

	l := model.RevokedCertificates{r}
	assert.NotNil(t, l.Find(123))
	assert.Nil(t, l.Find(23))
}

const testCrt = `-----BEGIN CERTIFICATE-----
MIICRTCCAcygAwIBAgIUUgP1I7qlkaVF0SuUd/HtShek77kwCgYIKoZIzj0EAwMw
TzELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t
MR4wHAYDVQQDDBVbVEVTVF0gVHJ1c3R5IFJvb3QgQ0EwHhcNMjEwNzAyMTU1ODAw
WhcNMjYwNzAxMTU1ODAwWjBSMQswCQYDVQQGEwJVUzELMAkGA1UEBxMCV0ExEzAR
BgNVBAoTCnRydXN0eS5jb20xITAfBgNVBAMMGFtURVNUXSBUcnVzdHkgTGV2ZWwg
MSBDQTB2MBAGByqGSM49AgEGBSuBBAAiA2IABMRI3GF/gDeC4IztWStz/Nkga2Vh
mKyZ1LK4Oe+T8/Xg9fnSzWMwvTMIlNV4cHfaiHT7GtBIYTATf7cvoQqcGMXvzU7C
2SHM7aK/zcViuVbu7t+wEUF4qSFDuSK3PH1GhqNmMGQwDgYDVR0PAQH/BAQDAgEG
MBIGA1UdEwEB/wQIMAYBAf8CAQEwHQYDVR0OBBYEFKT5ar2jVzVRzyKPB7/lTDKd
3S9ZMB8GA1UdIwQYMBaAFCpJu0kKGxPeP+wig4hJubVwAv+aMAoGCCqGSM49BAMD
A2cAMGQCMG5FNrmdZIMD8hWH9nVsUoLCQoLbuuvS/HdC8fPi5fwvdZvJtZ7K/JNG
IEN9D35UWQIwEsqs1R1K+zi6jfjBzuXCgKdvcOxRnxNOokh69FVCCoegVEDbgDBj
yMrvIi4tTwKn
-----END CERTIFICATE-----`
