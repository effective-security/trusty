package ca

import (
	"context"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/porto/xhttp/httperror"
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
	if in.ID != 0 {
		crt, err = s.db.GetCertificate(ctx, in.ID)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
	} else if in.IssuerSerial != nil {
		crt, err = s.db.GetCertificateByIKIDAndSerial(ctx, in.IssuerSerial.IKID, in.IssuerSerial.SerialNumber)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
	} else if len(in.SKID) > 0 {
		crts, err := s.db.GetCertificatesBySKID(ctx, in.SKID)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
		crt = crts[0]
	} else {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "invalid parameter")
	}

	revoked, err := s.db.RevokeCertificate(ctx, crt, time.Now().UTC(), int(in.Reason))
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to revoke certificate")
	}

	metricskey.CACertRevoked.IncrCounter(1, crt.IKID)

	s.publishCrlInBackground(crt.IKID)

	res := &pb.RevokedCertificateResponse{
		Revoked: revoked.ToDTO(),
	}
	return res, nil
}

// PublishCrls returns published CRLs
func (s *Service) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	return s.publishCrl(ctx, req.IKID)
}

// GetCRL returns the CRL
func (s *Service) GetCRL(ctx context.Context, in *pb.GetCrlRequest) (*pb.CrlResponse, error) {
	crl, err := s.db.GetCrl(ctx, in.IKID)
	if err == nil {
		return &pb.CrlResponse{
			Crl: crl.ToDTO(),
		}, nil
	}

	logger.ContextKV(ctx, xlog.TRACE,
		"ikid", in.IKID,
		"err", err.Error(),
	)

	resp, err := s.publishCrl(ctx, in.IKID)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to publish CRL")
	}

	res := &pb.CrlResponse{}

	if len(resp.Crls) > 0 {
		res.Crl = resp.Crls[0]
	}

	return res, nil
}

// SignOCSP returns OCSP response
func (s *Service) SignOCSP(ctx context.Context, in *pb.OCSPRequest) (*pb.OCSPResponse, error) {
	ocspRequest, err := ocsp.ParseRequest(in.Der)
	if err != nil ||
		ocspRequest.SerialNumber == nil {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "invalid request")
	}

	var ica *authority.Issuer
	if len(ocspRequest.IssuerKeyHash) > 0 {
		ica, err = s.ca.GetIssuerByKeyHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerKeyHash)
	} else if len(ocspRequest.IssuerNameHash) > 0 {
		ica, err = s.ca.GetIssuerByNameHash(ocspRequest.HashAlgorithm, ocspRequest.IssuerNameHash)
	} else {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "issuer not specified")
	}

	if err != nil {
		return nil, httperror.NewGrpcFromCtx(ctx, codes.NotFound, "issuer not found")
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
		return nil, httperror.WrapWithCtx(ctx, err, "unable to get revoked certificate")
	}

	if ri != nil {
		req.Status = authority.OCSPStatusRevoked
		req.Reason = ocsp.Unspecified
		req.RevokedAt = ri.RevokedAt.UTC()
	}

	logger.ContextKV(ctx, xlog.TRACE, "ikid", ikid, "serial", serial, "status", req.Status)

	der, err := ica.SignOCSP(req)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to sign OCSP")
	}

	metricskey.CAOcspSigned.IncrCounter(1, ikid, req.Status)

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
				RevocationTime: ri.RevokedAt.UTC(),
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
		ThisUpdate: xdb.Time(now),
		NextUpdate: xdb.Time(expiryTime),
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
				return res, httperror.WrapWithCtx(ctx, err, "failed to generate CRLs")
			}
			res.Crls = append(res.Crls, crl)

			_, err = s.publisher.PublishCRL(ctx, crl)
			if err != nil {
				return res, httperror.WrapWithCtx(ctx, err, "failed to publish CRLs")
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
