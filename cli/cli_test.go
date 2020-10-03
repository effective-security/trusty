package cli_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/trusty/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../"

func cmdAction(c ctl.Control, p interface{}) error {
	fmt.Fprintln(c.Writer(), "cmd executed!")
	return nil
}

func cmdClientAction(c ctl.Control, p interface{}) error {
	client := c.(*cli.Cli).Client()
	fmt.Fprintf(c.Writer(), "client: %T\n", client)
	return nil
}

func TestCLIDefaultHost(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test")
	app.UsageWriter(out)
	app.ErrorWriter(out)

	hostname, _ := os.Hostname()

	cli := cli.New(&ctl.ControlDefinition{
		App:    app,
		Output: out,
	}, cli.WithServer(hostname))
	cli.WithErrWriter(out)
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))

	assert.Panics(t, func() {
		cli.Client()
	})
	assert.Panics(t, func() {
		cli.CryptoProv()
	})

	out.Reset()
	cli.Parse([]string{"cliapp", "-D", "-V", "cmd", "client"})

	assert.Equal(t, hostname, cli.Server())
	assert.False(t, cli.IsJSON())
	assert.True(t, cli.Verbose())

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode(), "output: "+out.String())
	assert.Contains(t, out.String(), "client: *client.Client\n")
}

func TestCLIDefaultHostWithPort(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test")
	app.UsageWriter(out)
	app.ErrorWriter(out)

	hostname, _ := os.Hostname()
	serverURL := fmt.Sprintf("%s:7891", hostname)

	cli := cli.New(&ctl.ControlDefinition{
		App:    app,
		Output: out,
	}, cli.WithServer(serverURL))
	cli.WithErrWriter(out)
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))

	out.Reset()

	cli.Parse([]string{"cliapp", "cmd", "client"})
	assert.Equal(t, serverURL, cli.Server())

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode(), "output: "+out.String())
	assert.Contains(t, out.String(), "client: *client.Client\n")
}

func TestCLIEnsureClient(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test")
	app.UsageWriter(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:    app,
		Output: out,
	}, cli.WithServer(""), cli.WithTLS())
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))

	cert := "/tmp/trusty/certs/trusty_dev_peer.pem"
	key := "/tmp/trusty/certs/trusty_dev_peer-key.pem"

	require.Panics(t, func() {
		cli.Client()
	})

	cli.Parse([]string{"cliapp", "-s", "localhost", "-c", cert, "-k", key, "cmd", "client"})

	err := cli.EnsureClient()
	require.NoError(t, err)

	require.NotPanics(t, func() {
		cli.Client()
	})
	require.NotNil(t, cli.Client())

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode())
	assert.Contains(t, out.String(), "client: *client.Client\n")

	cli.WithClient(nil)
	require.Panics(t, func() {
		cli.Client()
	})
}

func TestCLIWithServiceCfgNoDefault(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test")
	app.UsageWriter(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:       app,
		Output:    out,
		ErrOutput: out,
	}, cli.WithServer(""))
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))

	out.Reset()
	cli.WithWriter(out).WithErrWriter(out)

	cli.Parse([]string{"cliapp", "cmd", "client"})

	assert.Panics(t, func() {
		cli.Config()
	})

	err := cli.EnsureServiceConfig()
	require.Error(t, err)

	assert.Contains(t, out.String(), "ERROR: use --server option\n")
}

func TestCLIWithServiceCfg_NotFound(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:       app,
		Output:    out,
		ErrOutput: out,
	}, cli.WithServer(""))
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		PreAction(cli.EnsureServiceConfig).
		Action(cli.RegisterAction(cmdClientAction, nil))

	out.Reset()
	cli.Parse([]string{"cliapp", "-s", "http://local", "cmd", "client"})

	assert.Panics(t, func() {
		cli.Config()
	})

	assert.Contains(t, out.String(), "specify --cfg option")
}

func TestCLIWithServiceCfg(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test")
	app.UsageWriter(out).
		Writer(out).
		ErrorWriter(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:    app,
		Output: out,
	}, cli.WithServiceCfg(), cli.WithServer(""))
	cli.WithErrWriter(out)
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))
	out.Reset()

	cfg, err := filepath.Abs(filepath.Join(projFolder, "etc/dev/trusty-config.json"))
	require.NoError(t, err)

	cli.Parse([]string{"cliapp", "cmd", "client", "--cfg", cfg})

	err = cli.EnsureServiceConfig()
	require.NoError(t, err)

	assert.NotNil(t, cli.Config())
	assert.Equal(t, cfg, cli.ConfigFlag())

	err = cli.EnsureCryptoProvider()
	require.NoError(t, err)

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode(), "output: "+out.String())
	assert.Contains(t, out.String(), "client: *client.Client\n")
}

func TestCLIWithHsmCfg(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test").
		UsageWriter(out).
		Writer(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:       app,
		Output:    out,
		ErrOutput: out,
	}, cli.WithHsmCfg(), cli.WithServer(""))
	cli.WithErrWriter(out)
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))
	out.Reset()

	cli.Parse([]string{"cliapp", "-s", "localhost", "cmd", "client", "--hsm-cfg", "/tmp/trusty/softhsm/unittest_hsm.json"})

	cli.WithCryptoProvider(nil)
	err := cli.EnsureCryptoProvider()
	require.NoError(t, err)
	// second time
	err = cli.EnsureCryptoProvider()
	require.NoError(t, err)

	prov, defCrypto := cli.CryptoProv()
	assert.NotNil(t, prov)
	assert.NotNil(t, defCrypto)

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode())
	assert.Contains(t, out.String(), "client: *client.Client\n")
}

func TestCLIWithHsmAndPlainText(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	app := ctl.NewApplication("cliapp", "test").
		UsageWriter(out).
		Writer(out)

	cli := cli.New(&ctl.ControlDefinition{
		App:       app,
		Output:    out,
		ErrOutput: out,
	}, cli.WithHsmCfg(), cli.WithServer(""))
	cli.WithErrWriter(out)
	defer cli.Close()

	cmd := app.Command("cmd", "Test command").
		PreAction(cli.PopulateControl)

	cmd.Command("client", "Test client").
		PreAction(cli.EnsureClient).
		Action(cli.RegisterAction(cmdClientAction, nil))
	out.Reset()

	cli.Parse([]string{"cliapp",
		"-s", "localhost",
		"cmd", "client",
		"--hsm-cfg", "/tmp/trusty/softhsm/unittest_hsm.json",
		"--plain-key",
	})

	cli.WithCryptoProvider(nil)
	err := cli.EnsureCryptoProvider()
	require.NoError(t, err)
	// second time
	err = cli.EnsureCryptoProvider()
	require.NoError(t, err)

	prov, defCrypto := cli.CryptoProv()
	assert.NotNil(t, prov)
	assert.NotNil(t, defCrypto)

	assert.Equal(t, ctl.RCOkay, cli.ReturnCode())
	assert.Contains(t, out.String(), "client: *client.Client\n")
}

func makeTestHandler(t *testing.T, expURI, responseBody string) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expURI, r.RequestURI, "received wrong URI")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, responseBody)
	}
	return http.HandlerFunc(h)
}

func TestReadStdin(t *testing.T) {
	_, err := cli.ReadStdin("")
	require.Error(t, err)
	assert.Equal(t, "empty file name", err.Error())

	b, err := cli.ReadStdin("-")
	assert.NoError(t, err)
	assert.Empty(t, b)
}
