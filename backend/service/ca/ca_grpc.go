package ca

import (
	"context"
	"strings"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/juju/errors"
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

// CreateCertificate returns the certificate
func (s *Service) CreateCertificate(context.Context, *pb.CreateCertificateRequest) (*pb.CertificateBundle, error) {
	return nil, errors.Errorf("not implemented")
}

// Issuers returns the issuing CAs
func (s *Service) Issuers(context.Context, *pb.EmptyRequest) (*pb.IssuersInfoResponse, error) {
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
