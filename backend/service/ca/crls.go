package ca

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/cloudflare/cfssl/helpers"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/juju/errors"
)

func (s *Service) createGenericCRL(ctx context.Context, issuer *authority.Issuer) (*pb.Crl, error) {
	bundle := issuer.Bundle()
	now := time.Now().UTC()
	expiryTime := now.Add(issuer.CrlExpiry())

	revokedInfoList, err := s.db.GetRevokedCertificatesByIssuer(ctx, issuer.SubjectKID())
	if err != nil {
		return nil, errors.Trace(err)
	}

	revokedCerts := make([]pkix.RevokedCertificate, len(revokedInfoList))
	for idx, ri := range revokedInfoList {
		// Parse the PEM encoded certificate
		cert, err := helpers.ParseCertificatePEM([]byte(ri.Certificate.Pem))
		if err != nil {
			return nil, errors.Trace(err)
		}

		revokedCerts[idx].SerialNumber = cert.SerialNumber
		revokedCerts[idx].RevocationTime = ri.RevokedAt
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
