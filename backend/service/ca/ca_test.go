package ca_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/go-phorce/trusty/backend/service/ca"
	"github.com/go-phorce/trusty/backend/trustymain"
	"github.com/go-phorce/trusty/backend/trustyserver"
	"github.com/go-phorce/trusty/backend/trustyserver/embed"
	"github.com/go-phorce/trusty/client"
	"github.com/go-phorce/trusty/tests/testutils"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *trustyserver.TrustyServer
	trustyClient *client.Client

	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]trustyserver.ServiceFactory{
	ca.ServiceName: ca.Factory,
}

var trueVal = true

func TestMain(m *testing.M) {
	var err error
	//	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")

	for i, httpCfg := range cfg.HTTPServers {
		switch httpCfg.Name {
		case "Health":
			cfg.HTTPServers[i].Disabled = &trueVal

		case "Trusty":
			cfg.HTTPServers[i].Services = []string{ca.ServiceName}
			cfg.HTTPServers[i].ListenURLs = []string{httpAddr}
		}
	}

	sigs := make(chan os.Signal, 2)

	app := trustymain.NewApp([]string{}).
		WithConfiguration(cfg).
		WithSignal(sigs)

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var rc int
	var expError error

	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if ret {
			trustyServer = app.Server("Trusty")
			trustyClient = embed.NewClient(trustyServer)

			// Run the tests
			rc = m.Run()

			trustyClient.Close()

			// trigger stop
			sigs <- syscall.SIGTERM
		}

	case <-time.After(20 * time.Second):
		break
	}

	// wait for stop
	wg.Wait()

	os.Exit(rc)
}

func TestIssuers(t *testing.T) {
	res, err := trustyClient.Authority.Issuers(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, res.Issuers)
}

func TestProfileInfo(t *testing.T) {
	_, err := trustyClient.Authority.ProfileInfo(context.Background(), nil)
	require.Error(t, err)
}

func TestCreateCertificate(t *testing.T) {
	_, err := trustyClient.Authority.CreateCertificate(context.Background(), nil)
	require.Error(t, err)
}
