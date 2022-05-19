package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/pkg/print"
	"github.com/pkg/errors"
)

// CaCmd is the parent for ca command
type CaCmd struct {
	Issuers        ListIssuersCmd      `cmd:"" help:"list issuers certificates"`
	Certs          ListCertsCmd        `cmd:"" help:"list certificates"`
	Revoked        ListRevokedCertsCmd `cmd:"" help:"list revoked certificates"`
	Profile        GetProfileCmd       `cmd:"" help:"show certificate profile"`
	Sign           SignCmd             `cmd:"" help:"sign certificate"`
	PublishCrl     PublishCrlsCmd      `cmd:"" help:"publish CRL"`
	Revoke         RevokeCmd           `cmd:"" help:"revoke certificate"`
	SetCertLabel   UpdateCertLabelCmd  `cmd:"" help:"set certificate label"`
	GetCertificate GetCertificateCmd   `cmd:"" help:"get certificate"`
}

// ListIssuersCmd shows issuers
type ListIssuersCmd struct {
	Limit  int64
	After  uint64
	Bundle bool
}

// Run the command
func (a *ListIssuersCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	res, err := client.CAClient().ListIssuers(context.Background(), &pb.ListIssuersRequest{
		Limit:  a.Limit,
		After:  a.After,
		Bundle: a.Bundle,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		print.Issuers(cli.Writer(), res.Issuers, a.Bundle)
	}
	return nil
}

// ListCertsCmd prints certificates
type ListCertsCmd struct {
	Ikid  string `kong:"arg" required:"" help:"Issuer key ID"`
	Limit int
	After string
}

// Run the command
func (a *ListCertsCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	after := uint64(0)
	if a.After != "" {
		after, err = db.ID(a.After)
		if err != nil {
			return errors.WithMessage(err, "unable to parse --after")
		}
	}

	res, err := client.CAClient().ListCertificates(context.Background(), &pb.ListByIssuerRequest{
		Ikid:  a.Ikid,
		Limit: int64(a.Limit),
		After: after,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		print.CertificatesTable(cli.Writer(), res.List)
	}

	return nil
}

// ListRevokedCertsCmd prints revoked certificates
type ListRevokedCertsCmd struct {
	Ikid  string `kong:"arg" required:"" help:"Issuer key ID"`
	Limit int
	After string
}

// Run the command
func (a *ListRevokedCertsCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	after := uint64(0)
	if a.After != "" {
		after, err = db.ID(a.After)
		if err != nil {
			return errors.WithMessage(err, "unable to parse --after")
		}
	}

	res, err := client.CAClient().ListRevokedCertificates(context.Background(), &pb.ListByIssuerRequest{
		Ikid:  a.Ikid,
		Limit: int64(a.Limit),
		After: after,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		print.RevokedCertificatesTable(cli.Writer(), res.List)
	}

	return nil
}

// GetProfileCmd shows the certifiate profile
type GetProfileCmd struct {
	Label string `kong:"arg" required:"" help:"Profile label"`
}

// Run the command
func (a *GetProfileCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	res, err := client.CAClient().ProfileInfo(context.Background(), &pb.CertProfileInfoRequest{
		Label: a.Label,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	cli.WriteJSON(res)
	// TODO: printer
	return nil
}

// SignCmd signs certificate request
type SignCmd struct {
	// Csr specifies CSR to sign
	Csr         string `required:"" help:"request file"`
	Profile     string `required:"" help:"profile name"`
	IssuerLabel string
	Token       string
	SAN         []string
	Label       string `help:"certificate label"`
	Out         string
}

// Run the command
func (a *SignCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	csr, err := cli.ReadFile(a.Csr)
	if err != nil {
		return errors.WithMessagef(err, "failed to load request")
	}

	res, err := client.CAClient().SignCertificate(context.Background(), &pb.SignCertificateRequest{
		RequestFormat: pb.EncodingFormat_PEM,
		Request:       csr,
		Profile:       a.Profile,
		IssuerLabel:   a.IssuerLabel,
		Label:         a.Label,
		Token:         a.Token,
		San:           a.SAN,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	pem := res.Certificate.Pem
	if !strings.HasSuffix(pem, "\n") {
		pem += "\n"
	}
	pem += res.Certificate.IssuersPem

	if a.Out != "" {
		err = ioutil.WriteFile(a.Out, []byte(pem), 0664)
		if err != nil {
			return errors.WithStack(err)
		}
	} else if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		fmt.Fprint(cli.Writer(), pem)
		fmt.Fprint(cli.Writer(), "\n")
	}

	return nil
}

// PublishCrlsCmd publish one or all CRLs
type PublishCrlsCmd struct {
	Ikid string
}

// Run the command
func (a *PublishCrlsCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	res, err := client.CAClient().PublishCrls(context.Background(), &pb.PublishCrlsRequest{
		Ikid: a.Ikid,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.IsJSON() {
		cli.WriteJSON(res)
	} else {
		print.CrlsTable(cli.Writer(), res.Clrs)
	}

	return nil
}

// RevokeCmd revokes a certifiate
type RevokeCmd struct {
	ID     uint64
	SKID   string
	IKID   string
	Serial string
	Reason int
}

// Run the command
func (a *RevokeCmd) Run(cli *Cli) error {
	if a.IKID == "" && a.Serial == "" && a.ID == 0 && a.SKID == "" {
		return errors.New("no certificate specified")
	}

	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	var is *pb.IssuerSerial
	if a.IKID != "" && a.Serial != "" {
		is = &pb.IssuerSerial{
			Ikid:         a.IKID,
			SerialNumber: a.Serial,
		}
	}

	res, err := client.CAClient().RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{
		Id:           a.ID,
		Skid:         a.SKID,
		IssuerSerial: is,
		Reason:       pb.Reason(a.Reason),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	cli.WriteJSON(res)

	// TODO: printer
	return nil
}

// UpdateCertLabelCmd allows to update certifiate label
type UpdateCertLabelCmd struct {
	ID    uint64 `kong:"arg" required:"" help:"certificate ID"`
	Label string
}

// Run the command
func (a *UpdateCertLabelCmd) Run(cli *Cli) error {
	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	res, err := client.CAClient().UpdateCertificateLabel(context.Background(), &pb.UpdateCertificateLabelRequest{
		Id:    a.ID,
		Label: a.Label,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	cli.WriteJSON(res)
	// TODO: printer
	return nil
}

// GetCertificateCmd specifies flags for the GetCertificate action
type GetCertificateCmd struct {
	ID     uint64
	SKID   string
	IKID   string
	Serial string
}

// Run the command
func (a *GetCertificateCmd) Run(cli *Cli) error {
	if a.IKID == "" && a.Serial == "" && a.ID == 0 && a.SKID == "" {
		return errors.New("no certificate specified")
	}

	client, err := cli.Client(config.CAServerName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer client.Close()

	var is *pb.IssuerSerial
	if a.IKID != "" && a.Serial != "" {
		is = &pb.IssuerSerial{
			Ikid:         a.IKID,
			SerialNumber: a.Serial,
		}
	}

	res, err := client.CAClient().GetCertificate(context.Background(), &pb.GetCertificateRequest{
		Id:           a.ID,
		Skid:         a.SKID,
		IssuerSerial: is,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	cli.WriteJSON(res)
	// TODO: printer
	return nil
}
