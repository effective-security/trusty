package ca

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/x/slices"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/csr"
	"google.golang.org/grpc/codes"
)

// SignCertificate returns the certificate
func (s *Service) SignCertificate(ctx context.Context, req *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	if req == nil || req.Profile == "" {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "missing profile")
	}
	if len(req.Request) == 0 {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "missing request")
	}

	var pemReq string

	switch req.RequestFormat {
	case pb.EncodingFormat_PEM:
		pemReq = string(req.Request)
	case pb.EncodingFormat_DER:
		b := bytes.NewBuffer([]byte{})
		_ = pem.Encode(b, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: req.Request})
		pemReq = b.String()
	default:
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "unsupported request_format: %v", req.RequestFormat)
	}

	var subj *csr.X509Subject
	if req.Subject != nil {
		subj = &csr.X509Subject{
			CommonName:   req.Subject.CommonName,
			Names:        make([]csr.X509Name, len(req.Subject.Names)),
			SerialNumber: req.Subject.SerialNumber,
		}
		for i, n := range req.Subject.Names {
			subj.Names[i] = csr.X509Name{
				Country:            n.Country,
				Province:           n.State,
				Locality:           n.Locality,
				Organization:       n.Organisation,
				OrganizationalUnit: n.OrganisationalUnit,
				SerialNumber:       n.SerialNumber,
			}
		}
	}

	var err error
	var ca *authority.Issuer
	if req.IssuerLabel != "" {
		ca, err = s.ca.GetIssuerByLabel(req.IssuerLabel)
		if err != nil {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.NotFound, "issuer not found: %s", req.IssuerLabel)
		}
	} else {
		ca, err = s.ca.GetIssuerByProfile(req.Profile)
		if err != nil {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.NotFound, "issuer not found for profile: %s", req.Profile)
		}
	}

	if ca.Profile(req.Profile) == nil {
		msg := fmt.Sprintf("%q issuer does not support the requested profile: %q", ca.Label(), req.Profile)
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, msg)
	}

	cr := csr.SignRequest{
		Request: pemReq,
		Profile: req.Profile,
		SAN:     req.SAN,
		Subject: subj,
	}
	for _, ex := range req.Extensions {
		cr.Extensions = append(cr.Extensions, csr.X509Extension{
			ID:       toOID(ex.ID),
			Critical: ex.Critical,
			Value:    ex.Value,
		})
	}

	if req.NotBefore != "" {
		cr.NotBefore = xdb.ParseTime(req.NotBefore).UTC()
	}
	if req.NotAfter != "" {
		cr.NotAfter = xdb.ParseTime(req.NotAfter).UTC()
	}

	cert, pem, err := ca.Sign(cr)
	if err != nil {
		logger.ContextKV(ctx, xlog.WARNING,
			"status", "failed to sign certificate",
			"err", err.Error())

		metricskey.CAFailSignCert.IncrCounter(1, ca.Label(), req.Profile)

		str := err.Error()
		if slices.ContainsString([]string{"invalid", "not allowed", "missing", "parse"}, str) {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "failed to sign certificate: %s", str)
		}
		return nil, httperror.WrapWithCtx(ctx, err, "failed to sign certificate")
	}

	metricskey.CACertIssued.IncrCounter(1, ca.Label(), req.Profile)

	mcert := model.NewCertificate(cert, req.OrgID, req.Profile, string(pem), ca.PEM(), req.Label, nil, req.Metadata)
	fn := mcert.FileName()
	mcert.Locations = append(mcert.Locations, s.cfg.RegistrationAuthority.Publisher.BaseURL+"/"+fn)

	mcert, err = s.db.RegisterCertificate(ctx, mcert)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR,
			"status", "failed to register certificate",
			"err", err.Error())

		if strings.Contains(err.Error(), "certificates_skid") {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.AlreadyExists, "the key was already used")
		}

		return nil, httperror.WrapWithCtx(ctx, err, "failed to register certificate")
	}

	if s.publisher != nil {
		_, err := s.publisher.PublishCertificate(context.Background(), mcert.ToPB(), fn)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"status", "failed to publish certificate",
				"err", err.Error())

			metricskey.CAFailPublishCert.IncrCounter(1, ca.Label())
			return nil, httperror.WrapWithCtx(ctx, err, "failed to publish certificate")
		}
	}

	logger.ContextKV(ctx, xlog.NOTICE,
		"status", "signed certificate",
		"id", mcert.ID,
		"subject", mcert.Subject,
		"label", mcert.Label,
		"locations", mcert.Locations,
		"meta", mcert.Metadata,
	)
	res := &pb.CertificateResponse{
		Certificate: mcert.ToPB(),
	}
	return res, nil
}

func toOID(s []int64) []int {
	size := len(s)
	oid := make([]int, size)
	for i := 0; i < size; i++ {
		oid[i] = int(s[i])
	}
	return oid
}
