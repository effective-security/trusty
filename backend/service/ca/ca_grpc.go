package ca

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xpki/authority"
	"google.golang.org/grpc/codes"
	"gopkg.in/yaml.v3"
)

// ProfileInfo returns the certificate profile info
func (s *Service) ProfileInfo(ctx context.Context, req *pb.CertProfileInfoRequest) (*pb.CertProfile, error) {
	if req == nil || req.Label == "" {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "missing label parameter")
	}

	var profile *authority.CertProfile

	ca, err := s.ca.GetIssuerByProfile(req.Label)
	if err == nil {
		profile = ca.Profile(req.Label)
	}
	if profile == nil {
		profile = s.ca.Profiles()[req.Label]
	}

	if profile == nil {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.NotFound, "profile not found: %s", req.Label)
	}

	return toCertProfilePB(profile, req.Label), nil
}

// GetIssuer returns the issuing CA
func (s *Service) GetIssuer(ctx context.Context, req *pb.IssuerInfoRequest) (*pb.IssuerInfo, error) {
	var issuer *authority.Issuer
	var err error
	if req.Label != "" {
		issuer, err = s.ca.GetIssuerByLabel(req.Label)
	} else if req.IKID != "" {
		issuer, err = s.ca.GetIssuerByKeyID(req.IKID)
	} else {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "either label or ikid are required")
	}

	if err != nil {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.NotFound, "issuer not found")
	}
	return issuerInfo(issuer, true), nil
}

// ListIssuers returns the issuing CAs
func (s *Service) ListIssuers(ctx context.Context, req *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error) {
	issuers := s.ca.Issuers()

	res := &pb.IssuersInfoResponse{
		Issuers: make([]*pb.IssuerInfo, 0, len(issuers)),
	}

	// TODO: pagination
	for _, issuer := range issuers {
		res.Issuers = append(res.Issuers, issuerInfo(issuer, req.Bundle))
	}

	return res, nil
}

func issuerInfo(issuer *authority.Issuer, withBundle bool) *pb.IssuerInfo {
	bundle := issuer.Bundle()
	ii := &pb.IssuerInfo{
		Certificate: bundle.CertPEM,
		Label:       issuer.Label(),
	}

	if withBundle {
		ii.Intermediates = bundle.CACertsPEM
		ii.Root = bundle.RootCertPEM
	}

	for name := range issuer.Profiles() {
		ii.Profiles = append(ii.Profiles, name)
	}
	return ii
}

// GetCertificate returns Certificate
func (s *Service) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	var crt *model.Certificate
	var err error
	if in.ID != 0 {
		crt, err = s.db.GetCertificate(ctx, in.ID)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
	} else {
		crts, err := s.db.GetCertificatesBySKID(ctx, in.SKID)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
		crt = crts[0]
	}
	// TODO: IssuerSerial
	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
	}
	return res, nil
}

// UpdateCertificateLabel returns the updated certificate
func (s *Service) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest) (*pb.CertificateResponse, error) {
	crt, err := s.db.UpdateCertificateLabel(ctx, req.ID, req.Label)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to update certificate")
	}
	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
	}
	return res, nil
}

// ListCertificates returns stream of Certificates
func (s *Service) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.ListCertificates(ctx, in.IKID, int(in.Limit), in.After)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to list certificates")
	}
	res := &pb.CertificatesResponse{
		Certificates: list.ToDTO(),
	}
	return res, nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (s *Service) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	list, err := s.db.ListRevokedCertificates(ctx, in.IKID, int(in.Limit), in.After)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to list certificates")
	}
	res := &pb.RevokedCertificatesResponse{
		RevokedCertificates: list.ToDTO(),
	}
	return res, nil
}

// ListOrgCertificates returns the Org certificates
func (s *Service) ListOrgCertificates(ctx context.Context, in *pb.ListOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.ListOrgCertificates(ctx, in.OrgID, int(in.Limit), in.After)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to get certificates")
	}
	res := &pb.CertificatesResponse{
		Certificates: list.ToDTO(),
	}
	return res, nil
}

// RegisterProfile registers the certificate profile
func (s *Service) RegisterProfile(ctx context.Context, req *pb.RegisterProfileRequest) (*pb.CertProfile, error) {
	var cfg = new(authority.CertProfile)
	err := yaml.Unmarshal(req.Config, cfg)
	if err != nil {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "unable to decode configuration: %s", err.Error())
	}

	isWildcard := cfg.IssuerLabel == "*"

	var issuer *authority.Issuer

	// check if profile is already served
	if !isWildcard {
		issuer, err = s.ca.GetIssuerByProfile(req.Label)
		if err == nil && issuer.Label() != cfg.IssuerLabel {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%q profile already served by %q issuer", req.Label, issuer.Label())
		}
		issuer, err = s.ca.GetIssuerByLabel(cfg.IssuerLabel)
		if err != nil {
			return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "issuer not found: %s", cfg.IssuerLabel)
		}
		issuer.AddProfile(req.Label, cfg)
	} else {
		s.ca.AddProfile(req.Label, cfg)
		for _, issuer := range s.ca.Issuers() {
			issuer.AddProfile(req.Label, cfg)
		}
	}

	_, err = s.db.RegisterCertProfile(ctx, &model.CertProfile{
		Label:       req.Label,
		IssuerLabel: cfg.IssuerLabel,
		Config:      string(req.Config),
	})
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to register profile: %s", err.Error())
	}

	return toCertProfilePB(cfg, req.Label), nil
}

func toCertProfilePB(cfg *authority.CertProfile, label string) *pb.CertProfile {
	p := &pb.CertProfile{
		Label:       label,
		IssuerLabel: cfg.IssuerLabel,
		Description: cfg.Description,
		Usages:      cfg.Usage,
		CAConstraint: &pb.CAConstraint{
			IsCA:       cfg.CAConstraint.IsCA,
			MaxPathLen: int32(cfg.CAConstraint.MaxPathLen),
		},
		OcspNoCheck:       cfg.OCSPNoCheck,
		Expiry:            cfg.Expiry.String(),
		Backdate:          cfg.Backdate.String(),
		AllowedExtensions: cfg.AllowedExtensionsStrings(),
		AllowedNames:      cfg.AllowedNames,
		AllowedDns:        cfg.AllowedDNS,
		AllowedEmail:      cfg.AllowedEmail,
		AllowedUri:        cfg.AllowedURI,
		PoliciesCritical:  cfg.PoliciesCritical,
		AllowedRoles:      cfg.AllowedRoles,
		DeniedRoles:       cfg.DeniedRoles,
	}
	if cfg.AllowedCSRFields != nil {
		p.AllowedFields = &pb.CSRAllowedFields{
			Subject: cfg.AllowedCSRFields.Subject,
			Dns:     cfg.AllowedCSRFields.DNSNames,
			Email:   cfg.AllowedCSRFields.EmailAddresses,
			Uri:     cfg.AllowedCSRFields.URIs,
			Ip:      cfg.AllowedCSRFields.IPAddresses,
		}
	}
	for _, pol := range cfg.Policies {
		pbPol := &pb.CertificatePolicy{
			ID: pol.ID.String(),
		}
		for _, q := range pol.Qualifiers {
			pbPol.Qualifiers = append(pbPol.Qualifiers, &pb.CertificatePolicyQualifier{
				Type:  q.Type,
				Value: q.Value,
			})
		}
		p.Policies = append(p.Policies, pbPol)
	}
	return p
}
