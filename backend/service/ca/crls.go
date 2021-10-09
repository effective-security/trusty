package ca

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"google.golang.org/grpc/codes"
)

func (s *Service) createGenericCRL(ctx context.Context, issuer *authority.Issuer) (*pb.Crl, error) {
	bundle := issuer.Bundle()
	now := time.Now().UTC()
	expiryTime := now.Add(issuer.CrlExpiry())

	revokedCerts := make([]pkix.RevokedCertificate, 0, 1000)
	last := uint64(0)
	for {
		revokedInfoList, err := s.db.ListRevokedCertificates(ctx, issuer.SubjectKID(), 0, last)
		if err != nil {
			return nil, errors.Trace(err)
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
		return nil, errors.Annotate(err, "failed to create CRL")
	}

	crl, err := x509.ParseCRL(crlBytes)
	if err != nil {
		return nil, errors.Trace(err)
	}

	mcrl, err := s.db.RegisterCrl(ctx, &model.Crl{
		IKID:       issuer.SubjectKID(),
		ThisUpdate: now,
		NextUpdate: expiryTime,
		Issuer:     bundle.Subject.String(),
		Pem:        string(pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: crlBytes})),
	})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to register CRL")
	}

	s.server.Audit(
		"CA",
		"CRLSigned",
		"",
		"",
		0,
		fmt.Sprintf("ikid=%s, issuer=%q, next_update='%v'",
			bundle.SubjectID,
			bundle.Cert.Subject.String(),
			crl.TBSCertList.NextUpdate.Format(time.RFC3339)),
	)

	return mcrl.ToDTO(), nil
}

func (s *Service) publishCrl(ctx context.Context, ikID string) (*pb.CrlsResponse, error) {
	logger.KV(xlog.INFO, "ikid", ikID)

	res := &pb.CrlsResponse{}
	for _, issuer := range s.ca.Issuers() {
		if ikID == "" || ikID == issuer.SubjectKID() {
			crl, err := s.createGenericCRL(ctx, issuer)
			if err != nil {
				logger.KV(xlog.ERROR,
					"ikid", issuer.SubjectKID(),
					"err", errors.Details(err),
				)
				return res, v1.NewError(codes.Internal, "failed to generate CRLs")
			}
			res.Clrs = append(res.Clrs, crl)

			_, err = s.publisher.PublishCRL(ctx, crl)
			if err != nil {
				logger.KV(xlog.ERROR,
					"ikid", issuer.SubjectKID(),
					"err", errors.Details(err),
				)
				return res, v1.NewError(codes.Internal, "failed to publish CRLs")
			}

			s.server.Audit(
				"CA",
				"CRLPublished",
				"",
				"",
				0,
				fmt.Sprintf("ikid=%s, issuer=%q, next_update='%v'",
					issuer.SubjectKID(),
					issuer.Bundle().Cert.Subject.String(),
					crl.NextUpdate.AsTime().Format(time.RFC3339)),
			)
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
				"err", errors.Details(err),
			)
		}
	}()
}
