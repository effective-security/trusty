package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/cli"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/rest/tlsconfig"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/juju/errors"
)

// AuthenticateFlags defines flags for Authenticate command
type AuthenticateFlags struct {
	NoBrowser *bool
	Provider  *string
}

// Authenticate starts authentication
func Authenticate(c ctl.Control, p interface{}) error {
	flags := p.(*AuthenticateFlags)
	cli := c.(*cli.Cli)

	srv := cli.Server()
	if srv == "" {
		return errors.New("please specify --server option")
	}

	hn, _ := os.Hostname()

	httpClient := retriable.New()

	if strings.HasPrefix(srv, "https://") {
		tlscfg, err := tlsconfig.NewClientTLSFromFiles("", "", cli.TLSCAFile())
		if err != nil {
			return errors.Annotate(err, "unable to build TLS configuration")
		}
		httpClient.WithTLS(tlscfg)
	}

	if flags.Provider == nil || *flags.Provider == "" {
		return errors.New("please specify --provider parameter")
	}

	res := new(v1.AuthStsURLResponse)
	path := fmt.Sprintf("%s?redirect_url=%s/v1/auth/done&device_id=%s&sts=%s", v1.PathForAuthURL, srv, hn, *flags.Provider)
	_, _, err := httpClient.Request(context.Background(), "GET", []string{srv}, path, nil, res)
	if err != nil {
		return errors.Trace(err)
	}

	if flags.NoBrowser == nil || !*flags.NoBrowser {
		execCommand := "xdg-open"
		if runtime.GOOS == "darwin" {
			execCommand = "open"
		}
		err = exec.Command(execCommand, res.URL).Start()
		if err != nil {
			return errors.Trace(err)
		}
		fmt.Fprintf(cli.Writer(), "opening auth URL in browser: %s\n", res.URL)
	} else {
		fmt.Fprintf(cli.Writer(), "open auth URL in browser:\n%s\n", res.URL)
	}

	return nil
}
