package cis_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/backend/trustymain"
	"github.com/ekspand/trusty/backend/trustyserver/embed"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	trustyClient *client.Client
)

const (
	projFolder = "../../../"
)

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
		case "Trusty-CIS":
			cfg.HTTPServers[i].Services = []string{cis.ServiceName}
			cfg.HTTPServers[i].ListenURLs = []string{httpAddr}

		default:
			cfg.HTTPServers[i].Disabled = &trueVal
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
			trustyServer = app.Server("Trusty-CIS")
			if trustyServer == nil {
				panic("Trusty-CIS not found!")
			}
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

func TestRoots(t *testing.T) {
	res, err := trustyClient.CertInfoService.Roots(context.Background(), &trustypb.EmptyRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Roots)
}
