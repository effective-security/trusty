package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/rest/tlsconfig"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/cli"
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
	noBrowser := flags.NoBrowser != nil && !*flags.NoBrowser

	var wg sync.WaitGroup
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		token, ok := r.URL.Query()["token"]
		if !ok || len(token) != 1 || token[0] == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing token parameter"))
			return
		}

		w.Header().Set(header.ContentType, header.TextPlain)
		fmt.Fprintf(w, "Authenticated! You can close the browser now.\n\nexport TRUSTY_AUTH_TOKEN=%s\n", token[0])

		dirname, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(cli.Writer(), "unable to determine home folder: %s", err.Error())
			return
		}

		folder := path.Join(dirname, ".config", "trusty")
		os.MkdirAll(folder, 0755)
		err = ioutil.WriteFile(path.Join(folder, ".token"), []byte(token[0]), 0600)
		if err != nil {
			fmt.Fprintf(cli.Writer(), "unable to store token: %s", err.Error())
			return
		}
	}
	http.HandleFunc("/", handler)

	if !noBrowser {
		wg.Add(1)
		go func() {
			log.Fatal(http.ListenAndServe(":38989", nil))
		}()
	}

	res := new(v1.AuthStsURLResponse)
	var path string
	if noBrowser {
		path = fmt.Sprintf("%s?redirect_url=%s/v1/auth/done&device_id=%s&sts=%s", v1.PathForAuthURL, srv, hn, *flags.Provider)
	} else {
		path = fmt.Sprintf("%s?redirect_url=http://localhost:38989&device_id=%s&sts=%s", v1.PathForAuthURL, hn, *flags.Provider)
	}
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
		wg.Wait()
	} else {
		fmt.Fprintf(cli.Writer(), "open auth URL in browser:\n%s\n", res.URL)
	}

	return nil
}
