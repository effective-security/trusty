package acme_test

import (
	"os"
	"testing"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/cli/acme"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsStorage(t *testing.T) {
	s, err := acme.NewAccountsStorage("", "https://localhost:7891", guid.MustCreate())
	require.NoError(t, err)

	rootdir := s.GetRootPath()
	require.NotEmpty(t, rootdir)

	defer os.RemoveAll(rootdir)

	err = s.ExistsAccountFilePath()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account.json: no such file or directory")

	pkey, err := s.GetPrivateKey()
	require.NoError(t, err)

	fingerprint, err := acmemodel.GetKeyFingerprint(pkey)
	require.NoError(t, err)
	assert.NotEmpty(t, fingerprint)

	_, err = s.LoadAccount(pkey)
	require.Error(t, err)

	jws := acme.NewJWS(pkey, "", nil)
	_, err = jws.SignContent("http://localhost", []byte{1, 2, 3, 4})
	require.NoError(t, err)

	_, err = jws.SignEABContent("http://localhost", "1234", []byte{1, 2, 3, 4})
	require.NoError(t, err)
}
