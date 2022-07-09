package certpublisher_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/effective-security/porto/x/fileutil"
	"github.com/effective-security/porto/x/guid"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/pkg/certpublisher"
	"github.com/effective-security/xpki/certutil"
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
		Pem:  testCRL,
		Ikid: guid.MustCreate(),
	})
	require.NoError(t, err)

	assert.NoError(t, fileutil.FileExists(fn2))
}

const testCRL = `-----BEGIN X509 CRL-----
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
hY1Ju9jQcvBiusz+GXK2yadW
-----END X509 CRL-----
`
