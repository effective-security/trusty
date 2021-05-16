package ca

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
)

// Issuers shows the Issuing CAs
func Issuers(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.AuthorityService.Issuers(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Issuers(c.Writer(), res.Issuers, true)
	}
	return nil
}

// GetProfileFlags defines flags for Profile command
type GetProfileFlags struct {
	Profile *string
	Label   *string
}

// Profile shows the certifiate profile
func Profile(c ctl.Control, p interface{}) error {
	flags := p.(*GetProfileFlags)
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	res, err := client.AuthorityService.ProfileInfo(context.Background(), &pb.CertProfileInfoRequest{
		Profile: *flags.Profile,
		Label:   *flags.Label,
	})
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), res)
	fmt.Fprint(c.Writer(), "\n")
	// TODO: printer
	return nil
}

// SignFlags specifies flags for the Sign action
type SignFlags struct {
	// Request specifies CSR to sign
	Request     *string
	Profile     *string
	IssuerLabel *string
	Token       *string
	SAN         *[]string
	Out         *string
}

// Sign certificate request
func Sign(c ctl.Control, p interface{}) error {
	flags := p.(*SignFlags)
	cli := c.(*cli.Cli)
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.Trace(err)
	}
	defer client.Close()

	csr, err := ioutil.ReadFile(*flags.Request)
	if err != nil {
		return errors.Annotatef(err, "failed to load request")
	}

	res, err := client.AuthorityService.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		RequestFormat: pb.EncodingFormat_PEM,
		Request:       string(csr),
		Profile:       *flags.Profile,
		IssuerLabel:   *flags.IssuerLabel,
		San:           *flags.SAN,
		Token:         *flags.Token,
		WithBundle:    true,
	})
	if err != nil {
		return errors.Trace(err)
	}

	if flags.Out != nil && *flags.Out != "" {
		err = ioutil.WriteFile(*flags.Out, []byte(res.Certificate), 0664)
		if err != nil {
			return errors.Trace(err)
		}
	} else if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		fmt.Fprint(c.Writer(), res.Certificate)
		fmt.Fprint(c.Writer(), "\n")
	}

	return nil
}
