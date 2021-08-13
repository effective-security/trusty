package trustymain

import (
	"os"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/payment"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../../"

var (
	testDirPath = filepath.Join(os.TempDir(), "tests", "trusty", guid.MustCreate())
)

func TestMain(m *testing.M) {
	_ = os.MkdirAll(testDirPath, 0700)
	defer os.RemoveAll(testDirPath)

	// Run stripe mocked backend
	payment.SetStripeMockedBackend()

	// Run the tests
	rc := m.Run()
	os.Exit(rc)
}

func Test_App_NoConfig(t *testing.T) {
	app := NewApp([]string{"--dry-run"})
	defer app.Close()

	err := app.Run(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration \"trusty-config.yaml\"")
	assert.Contains(t, err.Error(), "file \"trusty-config.yaml\" in [")
}

func Test_AppOnClose(t *testing.T) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c := &closer{}
	app := NewApp([]string{
		"--std",
		"--cfg", cfgFile,
		"--cis-listen-url", testutils.CreateURLs("http", "localhost"),
		"--wfe-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ca-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
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

func Test_AppInitWithRun(t *testing.T) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c := &closer{}
	app := NewApp([]string{
		"--dry-run",
		"--cfg", cfgFile,
		"--cis-listen-url", testutils.CreateURLs("http", "localhost"),
		"--wfe-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ca-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
	})

	err = app.Run(nil)
	assert.NoError(t, err)

	defer app.OnClose(c)
}

func Test_AppInitWithCfg(t *testing.T) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	cpuf := path.Join(testDirPath, "profiler")
	defer os.Remove(cpuf)

	c := &closer{}
	app := NewApp([]string{
		"--dry-run",
		"--cfg", cfgFile,
		"--cis-listen-url", testutils.CreateURLs("http", "localhost"),
		"--wfe-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ca-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
	})
	defer app.OnClose(c)

	err = app.loadConfig()
	require.NoError(t, err)

	// logs to file
	app.cfg.Logs.Directory = filepath.Join(testDirPath, "logs")
	err = app.initLogs()
	require.NoError(t, err)

	// logs to std
	app.cfg.Logs.Directory = ""
	err = app.initLogs()
	require.NoError(t, err)

	// TODO: initMetrics

	// CPU profiler
	err = app.initCPUProfiler(cpuf)
	require.NoError(t, err)

	_, err = app.containerFactory()
	require.NoError(t, err)
}

func Test_AppInstance_StartFailOnPort(t *testing.T) {
	cfgPath, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	require.NoError(t, err)

	listenURL := testutils.CreateURLs("http", "localhost")

	sigs := make(chan os.Signal, 2)
	app := NewApp([]string{
		"--std",
		"--cfg", cfgPath,
		"--cis-listen-url", listenURL,
		"--wfe-listen-url", listenURL,
		"--ca-listen-url", listenURL,
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
	}).WithSignal(sigs)
	defer app.Close()

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var expError error
	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			t.Log(expError.Error())
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if ret {
			t.Log("server started")
			// trigger stop
			sigs <- syscall.SIGUSR2
			sigs <- syscall.SIGTERM
		}

	case <-time.After(10 * time.Second):
		t.Log("timeout")
		break
	}

	// wait for stop
	wg.Wait()

	require.Error(t, expError)
	assert.Contains(t, expError.Error(), "bind: address already in use")
}

func Test_AppInstance_CryptoProvError(t *testing.T) {
	cfgPath, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	require.NoError(t, err)

	cfg, err := config.LoadConfig(cfgPath)
	require.NoError(t, err)

	sigs := make(chan os.Signal, 2)
	app := NewApp([]string{
		"--std",
		"--cfg", cfgPath,
		"--cis-listen-url", testutils.CreateURLs("http", "localhost"),
		"--wfe-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ca-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
		"--hsm-cfg", cfg.CryptoProv.Default,
		"--crypto-prov", cfg.CryptoProv.Default,
	}).WithSignal(sigs)
	defer app.Close()

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var expError error
	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			//t.Log(expError.Error())
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if ret {
			t.Log("server started")
			// trigger stop
			sigs <- syscall.SIGUSR2
			sigs <- syscall.SIGTERM
		}

	case <-time.After(10 * time.Second):
		break
	}

	// wait for stop
	wg.Wait()

	require.Error(t, expError)

	assert.Contains(t, expError.Error(), "could not build arguments for function")
	assert.Contains(t, expError.Error(), "duplicate provider specified for manufacturer: SoftHSM")
}

func Test_AppInstance_StartStop(t *testing.T) {
	cfgPath, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	require.NoError(t, err)

	sigs := make(chan os.Signal, 2)
	app := NewApp([]string{
		"--std",
		"--cfg", cfgPath,
		"--cis-listen-url", testutils.CreateURLs("http", "localhost"),
		"--wfe-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ca-listen-url", testutils.CreateURLs("http", "localhost"),
		"--ra-listen-url", testutils.CreateURLs("http", "localhost"),
	}).WithSignal(sigs)
	defer app.Close()

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var expError error
	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			t.Log(expError.Error())
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if assert.True(t, ret, "server NOT started") {
			t.Log("server started")
			// trigger stop
			sigs <- syscall.SIGUSR2
			sigs <- syscall.SIGTERM
		}

	case <-time.After(10 * time.Second):
		t.Log("failed to start")
		break
	}

	// wait for stop
	wg.Wait()

	require.NoError(t, expError)
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
