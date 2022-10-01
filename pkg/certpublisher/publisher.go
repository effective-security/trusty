package certpublisher

import (
	"context"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/pkg/storage"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/pkg", "certpublisher")

// Publisher interface
type Publisher interface {
	// PublishCertificate publishes issued cert
	PublishCertificate(context.Context, *pb.Certificate, string) (string, error)
	// PublishCRL publishes issued CRL
	PublishCRL(context.Context, *pb.Crl) (string, error)
}

type publisher struct {
	cfg *Config
}

// NewPublisher returns new Publisher
func NewPublisher(cfg *Config) (Publisher, error) {
	logger.KV(xlog.INFO, "cert_bucket", cfg.CertsBucket, "crl_bucket", cfg.CRLBucket)
	return &publisher{cfg}, nil
}

// PublishCertificate publishes issued cert
func (p *publisher) PublishCertificate(ctx context.Context, cert *pb.Certificate, filename string) (string, error) {
	location := fmt.Sprintf("%s/%s", p.cfg.CertsBucket, filename)

	logger.KV(xlog.INFO, "location", location)

	pem := strings.TrimSpace(cert.Pem)
	if len(cert.IssuersPem) > 0 {
		pem = pem + "\n" + strings.TrimSpace(cert.IssuersPem)
	}

	_, err := storage.WriteFile(ctx, location, []byte(pem))
	if err != nil {
		return "", errors.WithMessagef(err, "unable to write file to: "+location)
	}

	err = storage.SetContentType(ctx, location, "application/pem-certificate-chain")
	if err != nil {
		logger.ContextKV(ctx, xlog.WARNING, "reason", "SetContentType", "err", err.Error())
	}
	return location, nil
}

// PublishCRL publishes issued CRL
func (p *publisher) PublishCRL(ctx context.Context, crl *pb.Crl) (string, error) {
	fileName := fmt.Sprintf("%s/%s.crl", p.cfg.CRLBucket, crl.Ikid)

	logger.KV(xlog.INFO, "location", fileName)

	block, _ := pem.Decode([]byte(crl.Pem))
	if block == nil || block.Type != "X509 CRL" || len(block.Headers) != 0 {
		return "", errors.Errorf("unable to parse PEM CRL: block type %s", block.Type)
	}

	_, err := storage.WriteFile(ctx, fileName, block.Bytes)
	if err != nil {
		return "", errors.WithMessagef(err, "unable to write file to: "+fileName)
	}
	return fileName, nil
}
