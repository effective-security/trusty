package ca

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
)

// ProfileInfo returns the certificate profile info
func (s *Service) ProfileInfo(ctx context.Context, req *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	if req == nil || req.Profile == "" {
		return nil, v1.NewError(codes.InvalidArgument, "missing profile parameter")
	}

	ca, err := s.ca.GetIssuerByProfile(req.Profile)
	if err != nil {
		logger.Warningf("api=ProfileInfo, reason=no_issuer, profile=%q", req.Profile)
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
func (s *Service) SignCertificate(ctx context.Context, req *pb.SignCertificateRequest) (*pb.CertificateBundle, error) {
	if req == nil || req.Profile == "" {
		return nil, v1.NewError(codes.InvalidArgument, "missing profile")
	}
	if req.Request == "" {
		return nil, v1.NewError(codes.InvalidArgument, "missing request")
	}
	if req.RequestFormat != pb.EncodingFormat_PEM {
		return nil, v1.NewError(codes.InvalidArgument, "unsupported request_format: %v", req.RequestFormat)
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
		Request: req.Request,
		Profile: req.Profile,
		SAN:     req.San,
	}

	_, pem, err := ca.Sign(cr)
	if err != nil {
		return nil, v1.NewError(codes.Internal, "failed to sign certificate request: "+err.Error())
	}

	// TODO: metrics
	// TODO: Audit
	// TODO: Registration

	res := &pb.CertificateBundle{
		Certificate:   string(pem),
		Intermediates: ca.PEM(),
		Root:          ca.Bundle().RootCertPEM,
	}

	if req.WithBundle {
		res.Certificate += "\n" + ca.PEM()
	}

	return res, nil
}
