package trustymain

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../../"

var (
	nextPort = int32(0)
)

func Test_App_NoConfig(t *testing.T) {
	app := New([]string{"--dry-run"})
	defer app.Close()

	err := app.Run(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration \"trusty-config.json\"")
	assert.Contains(t, err.Error(), "file \"trusty-config.json\" in [")
}

func Test_AppOnClose(t *testing.T) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c := &closer{}
	app := New([]string{
		"--std",
		"--cfg", cfgFile,
		//"--listen-urls", createURLs("http","localhost"),
	})

	app.OnClose(c)
	assert.False(t, c.closed)

	err = app.loadConfig()
	require.NoError(t, err)

	err = app.Close()
	require.NoError(t, err)

	assert.True(t, c.closed)

	err = app.Close()
	require.Error(t, err)
	assert.Equal(t, "already closed", err.Error())
}

func Test_AppInit(t *testing.T) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c := &closer{}
	app := New([]string{
		"--std",
		"--dry-run",
		"--cfg", cfgFile,
		//"--listen-urls", createURLs("http","localhost"),
	})

	err = app.Run(nil)
	assert.NoError(t, err)

	defer app.OnClose(c)
}

type closer struct {
	closed bool
}

func (c *closer) Close() error {
	if c.closed {
		return errors.New("already closed")
	}
	c.closed = true
	return nil
}

// createURLs returns URL with a random port
func createURLs(scheme, host string) string {
	if nextPort == 0 {
		nextPort = 17891 + int32(rand.Intn(5000))
	}
	next := atomic.AddInt32(&nextPort, 1)
	return fmt.Sprintf("%s://%s:%d", scheme, host, next)
}
