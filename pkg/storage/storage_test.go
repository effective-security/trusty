package storage_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/effective-security/porto/x/fileutil"
	"github.com/effective-security/porto/x/guid"
	"github.com/effective-security/trusty/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDirPath = filepath.Join(os.TempDir(), "tests", "trusty", guid.MustCreate())
)

func TestMain(m *testing.M) {
	_ = os.MkdirAll(testDirPath, 0700)
	defer os.RemoveAll(testDirPath)

	// Run the tests
	rc := m.Run()
	os.Exit(rc)
}

func TestStorage(t *testing.T) {
	ctx := context.Background()

	data := []byte(`data`)
	fn := path.Join(testDirPath, guid.MustCreate())
	n, err := storage.WriteFile(ctx, fn, data)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.NoError(t, fileutil.FileExists(fn))

	data2, err := storage.ReadFile(ctx, fn)
	require.NoError(t, err)
	assert.Equal(t, data, data2)

	err = storage.DeletePath(ctx, fn)
	require.NoError(t, err)
}

func TestConnection(t *testing.T) {
	c, err := storage.ConnectionFromPath("gs://bucket")
	require.NoError(t, err)
	c.Close()
}
