package certpublisher

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/pkg/storage"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "certpublisher")

// Publisher interface
type Publisher interface {
	// PublishCertificate publishes issued cert
	PublishCertificate(context.Context, *pb.Certificate) (string, error)
	// PublishCRL publishes issued CRL
	PublishCRL(context.Context, *pb.Crl) (string, error)
}

type publisher struct {
	cfg *Config
}

// NewPublisher returns new Publisher
func NewPublisher(cfg *Config) (Publisher, error) {
	return &publisher{cfg}, nil
}

// PublishCertificate publishes issued cert
func (p *publisher) PublishCertificate(ctx context.Context, cert *pb.Certificate) (string, error) {
	sn := cert.SerialNumber[:12]
	n := new(big.Int)
	n, ok := n.SetString(cert.SerialNumber, 10)
	if ok {
		sn = base64.RawURLEncoding.EncodeToString(n.Bytes()[:9])
	}

	fileName := fmt.Sprintf("%s/%s/%s", p.cfg.CertsBucket, cert.Ikid[:4], sn)

	logger.KV(xlog.INFO, "location", fileName)

	_, err := storage.WriteFile(ctx, fileName, []byte(cert.Pem))
	if err != nil {
		return "", errors.Annotatef(err, "unable to write file to: "+fileName)
	}
	return fileName, nil
}

// PublishCRL publishes issued CRL
func (p *publisher) PublishCRL(ctx context.Context, crl *pb.Crl) (string, error) {
	fileName := fmt.Sprintf("%s/%s", p.cfg.CertsBucket, string(crl.Ikid))

	logger.KV(xlog.INFO, "location", fileName)

	_, err := storage.WriteFile(ctx, fileName, []byte(crl.Pem))
	if err != nil {
		return "", errors.Annotatef(err, "unable to write file to: "+fileName)
	}
	return fileName, nil
}
