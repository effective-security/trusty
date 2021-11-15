package ca

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/xlog"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/martinisecurity/trusty/pkg/csr"
	"google.golang.org/grpc/codes"
)

var (
	keyForCertIssued        = []string{"cert", "issued"}
	keyForCertRevoked       = []string{"cert", "revoked"}
	keyForCertSignFailed    = []string{"cert", "sign-failed"}
	keyForCertPublishFailed = []string{"cert", "publish-failed"}
	keyForCrlPublished      = []string{"crl", "published"}
	keyForCrlPublishFailed  = []string{"crl", "publish-failed"}
)

// SignCertificate returns the certificate
func (s *Service) SignCertificate(ctx context.Context, req *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	if req == nil || req.Profile == "" {
		return nil, v1.NewError(codes.InvalidArgument, "missing profile")
	}
	if len(req.Request) == 0 {
		return nil, v1.NewError(codes.InvalidArgument, "missing request")
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
		return nil, v1.NewError(codes.InvalidArgument, "unsupported request_format: %v", req.RequestFormat)
	}

	var subj *csr.X509Subject
	if req.Subject != nil {
		subj = &csr.X509Subject{
			CommonName: req.Subject.CommonName,
			Names:      make([]csr.X509Name, len(req.Subject.Names)),
		}
		for i, n := range req.Subject.Names {
			subj.Names[i] = csr.X509Name{
				C:  n.Country,
				ST: n.State,
				L:  n.Locality,
				O:  n.Organisation,
				OU: n.OrganisationalUnit,
			}
		}
	}

	var err error
	var ca *authority.Issuer
	if req.IssuerLabel != "" {
		ca, err = s.ca.GetIssuerByLabel(req.IssuerLabel)
		if err != nil {
			return nil, v1.NewError(codes.InvalidArgument, "issuer not found: %s", req.IssuerLabel)
		}
	} else {
		ca, err = s.ca.GetIssuerByProfile(req.Profile)
		if err != nil {
			return nil, v1.NewError(codes.InvalidArgument, "issuer not found for profile: %s", req.Profile)
		}
	}

	if ca.Profile(req.Profile) == nil {
		msg := fmt.Sprintf("%q issuer does not support the requested profile: %q", ca.Label(), req.Profile)
		return nil, v1.NewError(codes.InvalidArgument, msg)
	}

	cr := csr.SignRequest{
		Request: pemReq,
		Profile: req.Profile,
		SAN:     req.San,
		Subject: subj,
		// TODO:
		//Extensions: req.Extensions,
	}

	if req.NotBefore != nil {
		cr.NotBefore = req.NotBefore.AsTime()
	}
	if req.NotAfter != nil {
		cr.NotAfter = req.NotAfter.AsTime()
	}

	tags := []metrics.Tag{
		{Name: "profile", Value: req.Profile},
		{Name: "issuer", Value: ca.Label()},
	}

	cert, pem, err := ca.Sign(cr)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to sign certificate",
			"err", err)

		metrics.IncrCounter(keyForCertSignFailed, 1, tags...)
		return nil, v1.NewError(codes.Internal, "failed to sign certificate request")
	}

	metrics.IncrCounter(keyForCertIssued, 1, tags...)

	mcert := model.NewCertificate(cert, req.OrgId, req.Profile, string(pem), ca.PEM(), req.Label, nil, req.Metadata)
	fn := mcert.FileName()
	mcert.Locations = append(mcert.Locations, s.cfg.RegistrationAuthority.Publisher.BaseURL+"/"+fn)

	mcert, err = s.db.RegisterCertificate(ctx, mcert)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to register certificate",
			"err", err)

		if strings.Contains(err.Error(), "certificates_skid") {
			return nil, v1.NewError(codes.AlreadyExists, "the key was already used")
		}

		return nil, v1.NewError(codes.Internal, "failed to register certificate")
	}

	if s.publisher != nil {
		_, err := s.publisher.PublishCertificate(context.Background(), mcert.ToPB(), fn)
		if err != nil {
			logger.KV(xlog.ERROR,
				"status", "failed to publish certificate",
				"err", err)
			metrics.IncrCounter(keyForCertPublishFailed, 1, tags...)
			return nil, v1.NewError(codes.Internal, "failed to publish certificate")
		}
	}

	logger.KV(xlog.NOTICE,
		"status", "signed certificate",
		"id", mcert.ID,
		"subject", mcert.Subject,
	)
	res := &pb.CertificateResponse{
		Certificate: mcert.ToPB(),
	}
	return res, nil
}
