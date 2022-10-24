package ca

import (
	"context"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/effective-security/metrics"
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/porto/xhttp/pberror"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ocsp"
	"google.golang.org/grpc/codes"
)

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
		return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "invalid parameter")
	}
	if err != nil {
		logger.ContextKV(ctx, xlog.WARNING,
			"request", in,
			"err", err.Error(),
		)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to find certificate")
	}

	revoked, err := s.db.RevokeCertificate(ctx, crt, time.Now().UTC(), int(in.Reason))
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR,
			"request", in,
			"err", err.Error(),
		)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to revoke certificate")
	}

	tags := []metrics.Tag{
		{Name: "ikid", Value: crt.IKID},
		{Name: "serial", Value: crt.SerialNumber},
	}

	metrics.IncrCounter(metricskey.CACertRevoked, 1, tags...)

	s.publishCrlInBackground(crt.IKID)

	res := &pb.RevokedCertificateResponse{
		Revoked: revoked.ToDTO(),
	}
	return res, nil
}

// PublishCrls returns published CRLs
func (s *Service) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	return s.publishCrl(ctx, req.Ikid)
}

// GetCRL returns the CRL
func (s *Service) GetCRL(ctx context.Context, in *pb.GetCrlRequest) (*pb.CrlResponse, error) {
	crl, err := s.db.GetCrl(ctx, in.Ikid)
	if err == nil {
		return &pb.CrlResponse{
			Clr: crl.ToDTO(),
		}, nil
	}

	logger.ContextKV(ctx, xlog.TRACE,
		"ikid", in.Ikid,
		"err", err.Error(),
	)

	resp, err := s.publishCrl(ctx, in.Ikid)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR,
			"ikid", in.Ikid,
			"err", err.Error(),
		)
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to publish CRL")
	}

	res := &pb.CrlResponse{}

	if len(resp.Clrs) > 0 {
		res.Clr = resp.Clrs[0]
	}

	return res, nil
}

// SignOCSP returns OCSP response
func (s *Service) SignOCSP(ctx context.Context, in *pb.OCSPRequest) (*pb.OCSPResponse, error) {
	ocspRequest, err := ocsp.ParseRequest(in.Der)
	if err != nil ||
		ocspRequest.SerialNumber == nil {
		return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "invalid request")
	}

	var ica *authority.Issuer
	if len(ocspRequest.IssuerKeyHash) > 0 {
		ica, err = s.ca.GetIssuerByKeyHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerKeyHash)
	} else if len(ocspRequest.IssuerNameHash) > 0 {
		ica, err = s.ca.GetIssuerByNameHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerNameHash)
	} else {
		return nil, pberror.NewFromCtx(ctx, codes.InvalidArgument, "issuer not specified")
	}

	if err != nil {
		return nil, pberror.NewFromCtx(ctx, codes.NotFound, "issuer not found")
	}

	serial := ocspRequest.SerialNumber.String()

	req := &authority.OCSPSignRequest{
		SerialNumber: ocspRequest.SerialNumber,
		Status:       authority.OCSPStatusGood,
		IssuerHash:   ocspRequest.HashAlgorithm,
	}

	ikid := ica.Bundle().IssuerID
	ri, err := s.db.GetRevokedCertificateByIKIDAndSerial(ctx, ikid, serial)
	if err != nil && !xdb.IsNotFoundError(err) {
		logger.ContextKV(ctx, xlog.ERROR,
			"ikid", ikid, "serial", serial, "err", err.Error())
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to get revoked certificate")
	}

	if ri != nil {
		req.Status = authority.OCSPStatusRevoked
		req.Reason = ocsp.Unspecified
		req.RevokedAt = ri.RevokedAt
	}

	logger.ContextKV(ctx, xlog.TRACE, "ikid", ikid, "serial", serial, "status", req.Status)

	der, err := ica.SignOCSP(req)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR,
			"err", err.Error())
		return nil, pberror.NewFromCtx(ctx, codes.Internal, "unable to sign OCSP")
	}

	metrics.IncrCounter(metricskey.CAOcspSigned, 1,
		metrics.Tag{Name: "ikid", Value: ikid},
		metrics.Tag{Name: "status", Value: req.Status},
	)

	return &pb.OCSPResponse{Der: der}, nil
}

func (s *Service) createGenericCRL(ctx context.Context, issuer *authority.Issuer) (*pb.Crl, error) {
	bundle := issuer.Bundle()
	now := time.Now().UTC()
	expiryTime := now.Add(issuer.CrlExpiry())

	revokedCerts := make([]pkix.RevokedCertificate, 0, 1000)
	last := uint64(0)
	for {
		revokedInfoList, err := s.db.ListRevokedCertificates(ctx, issuer.SubjectKID(), 0, last)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(revokedInfoList) == 0 {
			break
		}

		for _, ri := range revokedInfoList {
			sn := new(big.Int)
			sn, _ = sn.SetString(ri.Certificate.SerialNumber, 10)
			revokedCerts = append(revokedCerts, pkix.RevokedCertificate{
				SerialNumber:   sn,
				RevocationTime: ri.RevokedAt,
			})
			last = ri.Certificate.ID
		}
	}

	crlBytes, err := bundle.Cert.CreateCRL(rand.Reader, issuer.Signer(), revokedCerts, now, expiryTime)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CRL")
	}

	mcrl, err := s.db.RegisterCrl(ctx, &model.Crl{
		IKID:       issuer.SubjectKID(),
		ThisUpdate: now,
		NextUpdate: expiryTime,
		Issuer:     bundle.Subject.String(),
		Pem:        string(pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: crlBytes})),
	})
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to register CRL")
	}

	return mcrl.ToDTO(), nil
}

func (s *Service) publishCrl(ctx context.Context, ikID string) (*pb.CrlsResponse, error) {
	logger.ContextKV(ctx, xlog.INFO,
		"ikid", ikID)

	res := &pb.CrlsResponse{}
	for _, issuer := range s.ca.Issuers() {
		if ikID == "" || ikID == issuer.SubjectKID() {
			crl, err := s.createGenericCRL(ctx, issuer)
			if err != nil {
				logger.ContextKV(ctx, xlog.ERROR,
					"ikid", issuer.SubjectKID(),
					"err", err.Error(),
				)
				return res, pberror.NewFromCtx(ctx, codes.Internal, "failed to generate CRLs")
			}
			res.Clrs = append(res.Clrs, crl)

			_, err = s.publisher.PublishCRL(ctx, crl)
			if err != nil {
				logger.ContextKV(ctx, xlog.ERROR,
					"ikid", issuer.SubjectKID(),
					"err", err.Error(),
				)
				return res, pberror.NewFromCtx(ctx, codes.Internal, "failed to publish CRLs")
			}
		}
	}

	return res, nil
}

func (s *Service) publishCrlInBackground(ikID string) {
	go func() {
		_, err := s.publishCrl(context.Background(), ikID)
		if err != nil {
			logger.KV(xlog.ERROR,
				"ikid", ikID,
				"err", err.Error(),
			)
		}
	}()
}
