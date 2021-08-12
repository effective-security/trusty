package certpublisher_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/pkg/certpublisher"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xpki/certutil"
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

func TestProvider(t *testing.T) {
	p, err := certpublisher.NewPublisher(&certpublisher.Config{
		CertsBucket: testDirPath,
		CRLBucket:   testDirPath,
	})
	require.NoError(t, err)
	require.NotNil(t, p)

	sn := new(big.Int)
	sn = sn.SetBytes(certutil.Random(20))

	ctx := context.Background()
	fn, err := p.PublishCertificate(ctx, &pb.Certificate{
		Pem:          "test",
		Skid:         guid.MustCreate(),
		Ikid:         guid.MustCreate(),
		SerialNumber: sn.String(),
	})
	require.NoError(t, err)

	assert.NoError(t, fileutil.FileExists(fn))

	fn2, err := p.PublishCRL(ctx, &pb.Crl{
		Pem:  "test",
		Ikid: guid.MustCreate(),
	})
	require.NoError(t, err)

	assert.NoError(t, fileutil.FileExists(fn2))
}
