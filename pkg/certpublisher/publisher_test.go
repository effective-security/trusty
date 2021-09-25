package certpublisher_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/internal/db/cadb/model"
	"github.com/martinisecurity/trusty/pkg/certpublisher"
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

	cm := &model.Certificate{
		Pem:          "test",
		SKID:         guid.MustCreate(),
		IKID:         guid.MustCreate(),
		SerialNumber: sn.String(),
	}

	fn, err := p.PublishCertificate(ctx, cm.ToPB(), cm.FileName())
	require.NoError(t, err)

	assert.NoError(t, fileutil.FileExists(fn))

	fn2, err := p.PublishCRL(ctx, &pb.Crl{
		Pem:  "test",
		Ikid: guid.MustCreate(),
	})
	require.NoError(t, err)

	assert.NoError(t, fileutil.FileExists(fn2))
}
