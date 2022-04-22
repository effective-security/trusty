package trustymain

import (
	"os"
	"path"
	"runtime/pprof"
	"testing"

	"github.com/effective-security/porto/x/guid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCpuProfileCloser(t *testing.T) {
	output := path.Join(os.TempDir(), "trusty", guid.MustCreate())
	os.MkdirAll(output, 0777)
	defer os.Remove(output)

	cpuf, err := os.Create(path.Join(output, "profiler"))
	require.NoError(t, err)

	pprof.StartCPUProfile(cpuf)
	closer := &cpuProfileCloser{}

	err = closer.Close()
	assert.NoError(t, err)
	err = closer.Close()
	assert.Error(t, err)
}
