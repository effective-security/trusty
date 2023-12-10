package print

import (
	"io"

	"github.com/effective-security/trusty/api/pb"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// JSON prints value to out
func JSON(w io.Writer, value interface{}) error {
	json, err := pb.MarshalJSON(value)
	if err != nil {
		return errors.WithMessage(err, "failed to encode")
	}
	_, _ = w.Write(json)
	_, _ = w.Write([]byte{'\n'})
	return nil
}

// Yaml prints value  to out
func Yaml(w io.Writer, value interface{}) error {
	y, err := yaml.Marshal(value)
	if err != nil {
		return errors.WithMessage(err, "failed to encode")
	}
	_, _ = w.Write(y)
	return nil
}

// Object prints value to out in format
func Object(w io.Writer, format string, value interface{}) error {
	if format == "yaml" {
		return Yaml(w, value)
	}
	if format == "json" {
		return JSON(w, value)
	}
	Print(w, value)
	return nil
}

// Print value
func Print(w io.Writer, value interface{}) {
	switch t := value.(type) {
	case *pb.ServerVersion:
		ServerVersion(w, t)
	case *pb.ServerStatusResponse:
		ServerStatusResponse(w, t)
	case *pb.CallerStatusResponse:
		CallerStatusResponse(w, t)
	case *pb.IssuersInfoResponse:
		Issuers(w, t.Issuers, false)
	case []*pb.IssuerInfo:
		Issuers(w, t, false)
	case *pb.RootsResponse:
		Roots(w, t.Roots, false)
	case []*pb.RootCertificate:
		Roots(w, t, false)
	case *pb.CertificatesResponse:
		CertificatesTable(w, t.Certificates)
	case []*pb.Certificate:
		CertificatesTable(w, t)
	case *pb.RevokedCertificatesResponse:
		RevokedCertificatesTable(w, t.RevokedCertificates)
	case []*pb.RevokedCertificate:
		RevokedCertificatesTable(w, t)
	case *pb.RevokedCertificateResponse:
		RevokedCertificate(w, t.Revoked, true)
	case *pb.RevokedCertificate:
		RevokedCertificate(w, t, true)
	case *pb.CrlsResponse:
		CrlsTable(w, t.Crls)
	case []*pb.Crl:
		CrlsTable(w, t)
	case *pb.Certificate:
		Certificate(w, t, true)
	case *pb.CertificateResponse:
		Certificate(w, t.Certificate, true)
	default:
		_ = JSON(w, value)
	}
}
