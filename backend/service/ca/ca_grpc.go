package ca

import (
	"context"

	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/juju/errors"
)

// ProfileInfo returns the certificate profile info
func (s *Service) ProfileInfo(context.Context, *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	return nil, errors.Errorf("not implemented")
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
