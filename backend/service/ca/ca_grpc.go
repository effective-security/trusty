package ca

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/xlog"
	"github.com/golang/protobuf/ptypes/empty"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/martinisecurity/trusty/pkg/csr"
	"golang.org/x/crypto/ocsp"
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
		Label:  ca.Label(),
		Issuer: ca.Bundle().CertPEM,
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

	ca, err := s.ca.GetIssuerByProfile(req.Profile)
	if err != nil {
		return nil, v1.NewError(codes.InvalidArgument, "issuer not found for profile: %s", req.Profile)
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

	mcert := model.NewCertificate(cert, req.OrgId, req.Profile, string(pem), ca.PEM(), req.Label, nil)
	fn := mcert.FileName()
	mcert.Locations = append(mcert.Locations, s.cfg.RegistrationAuthority.Publisher.BaseURL+"/"+fn)

	mcert, err = s.db.RegisterCertificate(ctx, mcert)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to register certificate",
			"err", err)

		if strings.Contains(err.Error(), "duplicate key") {
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

// PublishCrls returns published CRLs
func (s *Service) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	return s.publishCrl(ctx, req.Ikid)
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
	// TODO: IssuerSerial
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to find certificate")
	}
	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
	}
	return res, nil
}

// UpdateCertificateLabel returns the updated certificate
func (s *Service) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest) (*pb.CertificateResponse, error) {
	crt, err := s.db.UpdateCertificateLabel(ctx, req.Id, req.Label)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", req,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to update certificate")
	}
	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
	}
	return res, nil
}

// RevokeCertificate returns the revoked certificate
func (s *Service) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	var crt *model.Certificate
	var err error
	if in.Id != 0 {
		crt, err = s.db.GetCertificate(ctx, in.Id)
	} else if in.IssuerSerial != nil {
		crt, err = s.db.GetCertificateByIKIDAndSerial(ctx, in.IssuerSerial.Ikid, in.IssuerSerial.SerialNumber)
	} else if len(in.Skid) > 0 {
		crt, err = s.db.GetCertificateBySKID(ctx, in.Skid)
	} else {
		return nil, v1.NewError(codes.InvalidArgument, "invalid parameter")
	}
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to find certificate")
	}

	revoked, err := s.db.RevokeCertificate(ctx, crt, time.Now().UTC(), int(in.Reason))
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to revoke certificate")
	}

	tags := []metrics.Tag{
		{Name: "ikid", Value: crt.IKID},
		{Name: "serial", Value: crt.SerialNumber},
	}

	metrics.IncrCounter(keyForCertRevoked, 1, tags...)

	s.publishCrlInBackground(crt.IKID)

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
			"err", err,
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
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to list certificates")
	}
	res := &pb.RevokedCertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// GetOrgCertificates returns the Org certificates
func (s *Service) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.GetOrgCertificates(ctx, in.OrgId)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to get certificates")
	}
	res := &pb.CertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// GetCRL returns the CRL
func (s *Service) GetCRL(ctx context.Context, in *pb.GetCrlRequest) (*pb.CrlResponse, error) {
	crl, err := s.db.GetCrl(ctx, in.Ikid)
	if err == nil {
		return &pb.CrlResponse{
			Clr: crl.ToDTO(),
		}, nil
	}

	logger.KV(xlog.ERROR,
		"ikid", in.Ikid,
		"err", err,
	)

	resp, err := s.publishCrl(ctx, in.Ikid)
	if err != nil {
		logger.KV(xlog.ERROR,
			"ikid", in.Ikid,
			"err", err,
		)
		return nil, v1.NewError(codes.Internal, "unable to publish CRL")
	}

	res := &pb.CrlResponse{
		Clr: resp.Clrs[0],
	}

	return res, nil
}

// SignOCSP returns OCSP response
func (s *Service) SignOCSP(ctx context.Context, in *pb.OCSPRequest) (*pb.OCSPResponse, error) {
	ocspRequest, err := ocsp.ParseRequest(in.Der)
	if err != nil ||
		ocspRequest.SerialNumber == nil {
		return nil, v1.NewError(codes.InvalidArgument, "invalid request")
	}

	var ica *authority.Issuer
	if len(ocspRequest.IssuerKeyHash) > 0 {
		ica, err = s.ca.GetIssuerByKeyHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerKeyHash)
	} else if len(ocspRequest.IssuerNameHash) > 0 {
		ica, err = s.ca.GetIssuerByNameHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerNameHash)
	} else {
		return nil, v1.NewError(codes.InvalidArgument, "issuer not specified")
	}

	if err != nil {
		return nil, v1.NewError(codes.NotFound, "issuer not found")
	}

	serial := ocspRequest.SerialNumber.String()

	req := &authority.OCSPSignRequest{
		SerialNumber: ocspRequest.SerialNumber,
		Status:       "good",
		IssuerHash:   ocspRequest.HashAlgorithm,
	}

	ikid := ica.Bundle().IssuerID
	ri, err := s.db.GetRevokedCertificateByIKIDAndSerial(ctx, ikid, serial)
	if err != nil && !db.IsNotFoundError(err) {
		logger.KV(xlog.ERROR, "ikid", ikid, "serial", serial, "err", err)
		return nil, v1.NewError(codes.Internal, "unable to get revoked certificate")
	}

	if ri != nil {
		req.Status = "revoked"
		req.Reason = ocsp.Unspecified
		req.RevokedAt = ri.RevokedAt
	}

	logger.KV(xlog.TRACE, "ikid", ikid, "serial", serial, "status", req.Status)

	der, err := ica.SignOCSP(req)
	if err != nil {
		logger.KV(xlog.ERROR, "err", err)
		return nil, v1.NewError(codes.Internal, "unable to sign OCSP")
	}

	return &pb.OCSPResponse{Der: der}, nil
}
