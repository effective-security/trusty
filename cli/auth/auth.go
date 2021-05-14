package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

	res := new(v1.AuthStsURLResponse)
	path := fmt.Sprintf("/v1/auth/url?redirect_url=%s/v1/auth/done&device_id=%s", srv, hn)
	_, _, err := httpClient.Request(context.Background(), "GET", []string{srv}, path, nil, res)
	if err != nil {
		return errors.Trace(err)
	}

	if flags.NoBrowser == nil || !*flags.NoBrowser {
		err = exec.Command("xdg-open", res.URL).Start()
		if err != nil {
			return errors.Trace(err)
		}
		fmt.Fprintf(cli.Writer(), "openning auth URL in browser: %s\n", res.URL)
	} else {
		fmt.Fprintf(cli.Writer(), "open auth URL in browser:\n%s\n", res.URL)
	}

	return nil
}
