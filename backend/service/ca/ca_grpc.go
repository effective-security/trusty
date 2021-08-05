package ca

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/internal/db/cadb"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/xlog"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"google.golang.org/grpc/codes"
)

var (
	keyForCertIssued  = []string{"cert", "issued"}
	keyForCertRevoked = []string{"cert", "revoked"}
)

// ProfileInfo returns the certificate profile info
func (s *Service) ProfileInfo(ctx context.Context, req *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	if req == nil || req.Profile == "" {
		return nil, v1.NewError(codes.InvalidArgument, "missing profile parameter")
	}

	ca, err := s.ca.GetIssuerByProfile(req.Profile)
	if err != nil {
		logger.Warningf("reason=no_issuer, profile=%q", req.Profile)
		return nil, v1.NewError(codes.NotFound, "profile not found: %s", req.Profile)
	}

	label := strings.ToLower(req.Label)
	if label != "" && label != strings.ToLower(ca.Label()) {
		return nil, v1.NewError(codes.NotFound, "profile %q is served by %s issuer",
			req.Profile, ca.Label())
	}

	profile := ca.Profile(req.Profile)
	if profile == nil {
		return nil, v1.NewError(codes.NotFound, "%q issuer does not support the request profile: %q",
			ca.Label(), req.Profile)
	}

	res := &pb.CertProfileInfo{
		Issuer: ca.Label(),
		Profile: &pb.CertProfile{
			Description:       profile.Description,
			Usages:            profile.Usage,
			CaConstraint:      &pb.CAConstraint{},
			OcspNoCheck:       profile.OCSPNoCheck,
			Expiry:            profile.Expiry.String(),
			Backdate:          profile.Backdate.String(),
			AllowedExtensions: profile.AllowedExtensionsStrings(),
			AllowedNames:      profile.AllowedNames,
			AllowedDns:        profile.AllowedDNS,
			AllowedEmail:      profile.AllowedEmail,
		},
	}

	if profile.AllowedCSRFields != nil {
		res.Profile.AllowedFields = &pb.CSRAllowedFields{
			Subject: profile.AllowedCSRFields.Subject,
			Dns:     profile.AllowedCSRFields.DNSNames,
			Ip:      profile.AllowedCSRFields.IPAddresses,
			Email:   profile.AllowedCSRFields.EmailAddresses,
		}
	}

	return res, nil
}

// Issuers returns the issuing CAs
func (s *Service) Issuers(context.Context, *empty.Empty) (*pb.IssuersInfoResponse, error) {
	issuers := s.ca.Issuers()

	res := &pb.IssuersInfoResponse{
		Issuers: make([]*pb.IssuerInfo, len(issuers)),
	}

	for i, issuer := range issuers {
		bundle := issuer.Bundle()
		res.Issuers[i] = &pb.IssuerInfo{
			Certificate:   bundle.CertPEM,
			Intermediates: bundle.CACertsPEM,
			Root:          bundle.RootCertPEM,
			Label:         issuer.Label(),
		}
	}

	return res, nil
}

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
		pemReq = string(b.Bytes())
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
				O:  n.Organisation,
				OU: n.OrganisationalUnit,
			}
		}
	}

	ca, err := s.ca.GetIssuerByProfile(req.Profile)
	if err != nil {
		return nil, v1.NewError(codes.InvalidArgument, err.Error())
	}

	label := req.IssuerLabel
	if label != "" && label != ca.Label() {
		msg := fmt.Sprintf("%q issuer does not support the request profile: %q", label, req.Profile)
		return nil, v1.NewError(codes.InvalidArgument, msg)

	}

	cr := csr.SignRequest{
		Request: pemReq,
		Profile: req.Profile,
		SAN:     req.San,
		Subject: subj,
		// TODO:
		//Extensions: req.E,
	}

	cert, pem, err := ca.Sign(cr)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to sign certificate",
			"err", errors.Details(err))
		return nil, v1.NewError(codes.Internal, "failed to sign certificate request")
	}

	tags := []metrics.Tag{
		{Name: "profile", Value: req.Profile},
		{Name: "issuer", Value: ca.Label()},
	}

	metrics.IncrCounter(keyForCertIssued, 1, tags...)

	mcert := model.NewCertificate(cert, req.OrgId, req.Profile, string(pem), ca.PEM())
	mcert, err = s.db.RegisterCertificate(ctx, mcert)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to register certificate",
			"err", errors.Details(err))

		return nil, v1.NewError(codes.Internal, "failed to register certificate")
	}
	logger.KV(xlog.NOTICE,
		"status", "signed certificate",
		"id", mcert.ID,
		"subject", mcert.Subject,
	)

	res := &pb.CertificateResponse{
		Certificate: mcert.ToDTO(),
	}

	return res, nil
}

// PublishCrls returns published CRLs
func (s *Service) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	logger.KV(xlog.TRACE, "ikid", req.Ikid)

	res := &pb.CrlsResponse{}
	for _, issuer := range s.ca.Issuers() {
		if req.Ikid == "" || req.Ikid == issuer.SubjectKID() {
			crl, err := s.createGenericCRL(ctx, issuer)
			if err != nil {
				logger.KV(xlog.ERROR,
					"issuer_id", issuer.SubjectKID(),
					"err", errors.Details(err),
				)
				return nil, v1.NewError(codes.Internal, "failed to publish CRLs")
			}
			res.Clrs = append(res.Clrs, crl)
		}
	}
	return res, nil
}

// GetCertificate returns Certificate
func (s *Service) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	var crt *model.Certificate
	var err error
	if in.Id != 0 {
		crt, err = s.db.GetCertificate(ctx, in.Id)
	} else {
		crt, err = s.db.GetCertificateBySKID(ctx, in.Skid)
	}
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to find certificate")
	}
	res := &pb.CertificateResponse{
		Certificate: crt.ToDTO(),
	}
	return res, nil
}

// GetOrgCertificates returns the Org certificates
func (s *Service) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.GetOrgCertificates(ctx, in.OrgId)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to get certificates")
	}
	res := &pb.CertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// RevokeCertificate returns the revoked certificate
func (s *Service) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	var crt *model.Certificate
	var err error
	if in.Id != 0 {
		crt, err = s.db.GetCertificate(ctx, in.Id)
	} else {
		crt, err = s.db.GetCertificateBySKID(ctx, in.Skid)
	}
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to find certificate")
	}

	revoked, err := s.db.RevokeCertificate(ctx, crt, time.Now().UTC(), int(in.Reason))
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to revoke certificate")
	}

	res := &pb.RevokedCertificateResponse{
		Revoked: revoked.ToDTO(),
	}
	return res, nil
}

// ListCertificates returns stream of Certificates
func (s *Service) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.ListCertificates(ctx, in.Ikid, int(in.Limit), in.After)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to list certificates")
	}
	res := &pb.CertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (s *Service) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	list, err := s.db.ListRevokedCertificates(ctx, in.Ikid, int(in.Limit), in.After)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to list certificates")
	}
	res := &pb.RevokedCertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// Db returns DB
// Used in Unittests
func (s *Service) Db() cadb.CaDb {
	return s.db
}
