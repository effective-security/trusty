package ca

import (
	"context"
	"fmt"
	"time"

	"github.com/effective-security/porto/xhttp/pberror"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/effective-security/xpki/csr"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"gopkg.in/yaml.v2"
)

// ListDelegatedIssuers returns the delegated issuing CAs
func (s *Service) ListDelegatedIssuers(ctx context.Context, req *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error) {
	list, err := s.db.ListIssuers(ctx, int(req.Limit), req.After)
	if err != nil {
		logger.KV(xlog.ERROR, "request", req, "err", err)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to list issuers")
	}

	res := &pb.IssuersInfoResponse{}

	// TODO: pagination
	for _, issuer := range list {
		var cfg = new(authority.IssuerConfig)
		err := yaml.Unmarshal([]byte(issuer.Config), cfg)
		if err != nil {
			logger.KV(xlog.ERROR, "issuer", issuer, "err", err)
			return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to decode configuration")
		}

		ii := &pb.IssuerInfo{
			Label:       issuer.Label,
			Type:        "delegated",
			Status:      pb.IssuerStatus(issuer.Status),
			Certificate: cfg.CertFile,
		}
		if req.Bundle {
			ii.Intermediates = cfg.CABundleFile
			ii.Root = cfg.RootBundleFile
		}
		res.Issuers = append(res.Issuers, ii)
	}

	return res, nil
}

// ArchiveDelegatedIssuer archives a delegated issuer.
func (s *Service) ArchiveDelegatedIssuer(ctx context.Context, req *pb.IssuerInfoRequest) (*pb.IssuerInfo, error) {
	return nil, nil
}

// RegisterDelegatedIssuer creates new delegate issuer.
func (s *Service) RegisterDelegatedIssuer(ctx context.Context, req *pb.SignCertificateRequest) (*pb.IssuerInfo, error) {
	if req.Label == "" || req.Profile == "" || len(req.Request) > 0 || req.OrgId == 0 {
		return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "invalid request")
	}

	if s.cfg.DelegatedIssuers.GetDisabled() {
		return nil, pberror.NewFromCtx(ctx, codes.Unimplemented, "delegated issuers not allowed")
	}

	iss, err := s.ca.GetIssuerByLabel(req.Label)
	if err == nil && iss != nil {
		return nil, pberror.NewFromCtx(ctx, codes.AlreadyExists, "issuer already registered with this label")
	}

	// ensure issuer exists pefore creating a key
	iss, err = s.ca.GetIssuerByProfile(req.Profile)
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.NotFound, "issuer not found for profile: %s", req.Profile)
	}

	delegatedIssuerLabel := fmt.Sprintf("%s%d", s.cfg.DelegatedIssuers.IssuerLabelPrefix, req.OrgId)
	profiles, err := s.db.GetCertProfilesByIssuer(ctx, delegatedIssuerLabel)
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to load profiles: %s", err.Error())
	}

	now := time.Now()
	keyLabel := fmt.Sprintf("%s-delegated-%d-%02d%02d%02d-%02d%02d",
		s.cfg.ClusterName, req.OrgId, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

	crypto, err := s.delegatedCrypto()
	if err != nil {
		logger.KV(xlog.ERROR, "err", err)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to load crypto provider")
	}
	prov := csr.NewProvider(crypto)
	sreq := prov.NewSigningCertificateRequest(keyLabel, "ECDSA", 256, "", nil, nil)

	csrPEM, keyBytes, _, _, err := prov.CreateRequestAndExportKey(sreq)
	if err != nil {
		logger.KV(xlog.ERROR, "err", err)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "failed to create key")
	}

	req.Request = csrPEM
	req.RequestFormat = pb.EncodingFormat_PEM

	signRes, err := s.SignCertificate(ctx, req)
	if err != nil {
		logger.KV(xlog.ERROR, "err", err)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "failed to create key")
	}

	cfg := &authority.IssuerConfig{
		Label:           delegatedIssuerLabel,
		Type:            "delegated",
		CertFile:        signRes.Certificate.Pem,
		KeyFile:         string(keyBytes),
		CABundleFile:    signRes.Certificate.IssuersPem,
		RootBundleFile:  iss.Bundle().RootCertPEM,
		AIA:             s.cfg.DelegatedIssuers.AIA,
		AllowedProfiles: s.cfg.DelegatedIssuers.AllowedProfiles,
		Profiles:        make(map[string]*authority.CertProfile),
	}

	signer, err := s.ca.Crypto().NewSignerFromPEM(keyBytes)
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "unable to create signer from private key: %s", err.Error())
	}

	for _, p := range profiles {
		var profile = new(authority.CertProfile)
		err := yaml.Unmarshal([]byte(p.Config), profile)
		if err != nil {
			return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "unable to decode profile: %s", err.Error())
		}

		cfg.Profiles[p.Label] = profile
	}

	for name, profile := range s.ca.Profiles() {
		if profile.IssuerLabel == "*" {
			cfg.Profiles[name] = profile
		}
	}

	issuer, err := authority.CreateIssuer(cfg,
		[]byte(cfg.CertFile),
		certutil.JoinPEM([]byte(cfg.CABundleFile), s.ca.CaBundle),
		certutil.JoinPEM([]byte(cfg.RootBundleFile), s.ca.RootBundle),
		signer,
	)
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "failed to create issuer: %s", err.Error())
	}

	err = s.ca.AddIssuer(issuer)
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "failed to add issuer: %s", err.Error())
	}

	jsoncfg, _ := yaml.Marshal(cfg)
	_, err = s.db.RegisterIssuer(ctx, &model.Issuer{
		Label:  cfg.Label,
		Status: int(pb.IssuerStatus_ACTIVE),
		Config: string(jsoncfg),
	})
	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "failed to save issuer: %s", err.Error())
	}

	return issuerInfo(issuer, true), nil
}

func (s *Service) delegatedCrypto() (cryptoprov.Provider, error) {
	if s.cfg.DelegatedIssuers.CryptoProvider != "" {
		prov, err := s.ca.Crypto().ByManufacturer(
			s.cfg.DelegatedIssuers.CryptoProvider,
			s.cfg.DelegatedIssuers.CryptoModel)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to load crypto provider %s, model %q",
				s.cfg.DelegatedIssuers.CryptoProvider,
				s.cfg.DelegatedIssuers.CryptoModel)
		}
		return prov, nil
	}
	return s.ca.Crypto().Default(), nil
}
