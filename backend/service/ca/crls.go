package ca

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/juju/errors"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/internal/db/cadb/model"
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
		Pem:        base64.StdEncoding.EncodeToString(crlBytes),
	})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to register CRL")
	}

	s.server.Audit(
		"CA",
		"CRLPublished",
		"",
		"",
		0,
		fmt.Sprintf("issuer_id=%s, issuer=%q, next_update='%v'",
			bundle.SubjectID,
			bundle.Cert.Subject.String(),
			crl.TBSCertList.NextUpdate.Format(time.RFC3339)),
	)

	return mcrl.ToDTO(), nil
}
