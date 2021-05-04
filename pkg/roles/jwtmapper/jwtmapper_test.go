package jwtmapper_test

import (
	"net/http"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/pkg/roles"
	"github.com/ekspand/trusty/pkg/roles/jwtmapper"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	_, err := jwtmapper.LoadConfig("testdata/missing.json")
	require.Error(t, err)
	assert.Equal(t, "open testdata/missing.json: no such file or directory", err.Error())

	_, err = jwtmapper.LoadConfig("testdata/roles_corrupted.1.json")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal JSON: "testdata/roles_corrupted.1.json": invalid character 'v' looking for beginning of value`, err.Error())

	_, err = jwtmapper.LoadConfig("testdata/roles_corrupted.yaml")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal YAML: "testdata/roles_corrupted.yaml": yaml: line 4: did not find expected alphabetic or numeric character`, err.Error())

	_, err = jwtmapper.LoadConfig("testdata/roles_corrupted.2.json")
	require.Error(t, err)
	assert.Equal(t, `missing kid: "testdata/roles_corrupted.2.json"`, err.Error())

	_, err = jwtmapper.LoadConfig("testdata/roles_no_kid.json")
	require.Error(t, err)
	assert.Equal(t, `missing kid: "testdata/roles_no_kid.json"`, err.Error())

	_, err = jwtmapper.LoadConfig("testdata/roles_no_keys.json")
	require.Error(t, err)
	assert.Equal(t, `missing keys: "testdata/roles_no_keys.json"`, err.Error())

	cfg, err := jwtmapper.LoadConfig("testdata/roles.json")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Keys))
	assert.Equal(t, "1", cfg.KeyID)
	assert.Equal(t, "trusty", cfg.Issuer)
	assert.Equal(t, "trusty", cfg.Audience)
	//assert.Equal(t, roles.trustyClient, cfg.DefaultRole)
	//assert.Equal(t, 2, len(cfg.RolesMap[roles.trustyAdmin]))
}

func Test_Load(t *testing.T) {
	_, err := jwtmapper.Load("testdata/missing.json")
	require.Error(t, err)
	assert.Equal(t, "open testdata/missing.json: no such file or directory", err.Error())

	_, err = jwtmapper.Load("testdata/roles_corrupted.1.json")
	require.Error(t, err)

	_, err = jwtmapper.Load("testdata/roles_corrupted.2.json")
	require.Error(t, err)

	m, err := jwtmapper.Load("testdata/roles.json")
	require.NoError(t, err)
	id, key := m.CurrentKey()
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, key)

	m, err = jwtmapper.Load("testdata/roles.yaml")
	require.NoError(t, err)
	id, key = m.CurrentKey()
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, key)

	_, err = jwtmapper.Load("")
	require.NoError(t, err)
}

func Test_Sign(t *testing.T) {
	p, err := jwtmapper.Load("testdata/roles.json")
	require.NoError(t, err)
	p1, err := jwtmapper.Load("testdata/roles.1.json")
	require.NoError(t, err)
	p2, err := jwtmapper.Load("testdata/roles.2.json")
	require.NoError(t, err)

	t.Run("default role", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "daniel@ekspand.com",
		}

		auth, err := p.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device123")
		assert.True(t, p.Applicable(r))

		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Equal(t, roles.TrustyClient, id.Role())
		assert.Equal(t, userInfo.Email, id.Name())
		assert.Equal(t, "123", id.UserID())
	})

	t.Run("admin role", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "denis@ekspand.com",
		}

		auth, err := p.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device123")
		assert.True(t, p.Applicable(r))

		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Equal(t, roles.TrustyAdmin, id.Role())
		assert.Equal(t, userInfo.Email, id.Name())
	})

	t.Run("invalid_sig", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "denis@ekspand.com",
		}

		auth, err := p.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device123")

		assert.True(t, p2.Applicable(r))
		_, err = p2.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, "failed to verify token: signature is invalid", err.Error())
	})

	t.Run("invalid_issuer", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "denis@ekspand.com",
		}

		auth, err := p.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device123")

		assert.True(t, p1.Applicable(r))
		_, err = p1.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, "invalid issuer: trusty", err.Error())
	})
}

func Test_Verify(t *testing.T) {
	p, err := jwtmapper.Load("testdata/roles.json")
	require.NoError(t, err)

	userInfo := &v1.UserInfo{
		ID:    "123",
		Email: "denis@ekspand.com",
	}
	auth, err := p.SignToken(userInfo, "device123", time.Second)
	require.NoError(t, err)

	t.Run("invalid_token", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken+"123", "device123")
		_, err = p.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, "failed to verify token: signature is invalid", err.Error())
	})

	t.Run("invalid_deviceID", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device456")
		_, err = p.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, "invalid deviceID: device456", err.Error())
	})

	t.Run("not_applicable_header", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", auth.AccessToken)
		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Nil(t, id)
	})
}

// setAuthorizationHeader applies Authorization header
func setAuthorizationHeader(r *http.Request, token, deviceID string) {
	r.Header.Set(header.Authorization, header.Bearer+" "+token)
	if deviceID != "" {
		r.Header.Set(header.XDeviceID, deviceID)
	}
}
