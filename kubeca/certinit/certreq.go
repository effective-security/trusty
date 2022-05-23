package certinit

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"strings"
	"time"

	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/cryptoprov/inmemcrypto"
	"github.com/effective-security/xpki/csr"
	"github.com/effective-security/xpki/x/print"
	"github.com/pkg/errors"
	capi "k8s.io/api/certificates/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	randAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	randAlphabetLength = big.NewInt(int64(len(randAlphabet)))
)

var usages = map[string][]capi.KeyUsage{
	"peer": {
		capi.UsageDigitalSignature,
		capi.UsageKeyEncipherment,
		capi.UsageServerAuth,
		capi.UsageClientAuth,
	},
	"client": {
		capi.UsageDigitalSignature,
		capi.UsageKeyEncipherment,
		capi.UsageClientAuth,
	},
	"server": {
		capi.UsageDigitalSignature,
		capi.UsageKeyEncipherment,
		capi.UsageServerAuth,
	},
}

func (r *Request) requestCertificate(ctx context.Context, client MinCertificates) error {
	issuerAndProfile := strings.Split(r.SignerName, "/")
	if len(issuerAndProfile) != 2 {
		return errors.Errorf("unsupported signer: " + r.SignerName)
	}

	profileUsages := usages[issuerAndProfile[1]]
	if profileUsages == nil {
		return errors.Errorf("unsupported profile: " + r.SignerName)
	}

	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := csr.CertificateRequest{
		KeyRequest: prov.NewKeyRequest("", "ECDSA", 256, csr.SigningKey),
		SAN:        r.san,
	}

	pemCsr, pemKeyBytes, _, _, err := prov.CreateRequestAndExportKey(&req)
	if err != nil {
		return errors.WithStack(err)
	}

	keyFile := path.Join(r.CertDir, "tls.key")
	// TODO: for 0600 the POD got access denied
	if err := ioutil.WriteFile(keyFile, pemKeyBytes, 0644); err != nil {
		return errors.WithMessage(err, "unable to save key")
	}

	logger.KV(xlog.INFO, "status", "wrote_key", "file", keyFile)

	csrFile := path.Join(r.CertDir, "tls.csr")
	if err := ioutil.WriteFile(csrFile, pemCsr, 0644); err != nil {
		return errors.WithMessage(err, "unable to save CSR")
	}

	logger.KV(xlog.INFO, "status", "wrote_csr", "file", csrFile)

	// Submit a certificate signing request, wait for it to be approved, then save
	// the signed certificate to the file system.
	certificateSigningRequestName := r.requestName()
	certificateSigningRequest := &capi.CertificateSigningRequest{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:   certificateSigningRequestName,
			Labels: r.labelsMap,
		},
		Spec: capi.CertificateSigningRequestSpec{
			Request:    pemCsr,
			Usages:     profileUsages,
			SignerName: r.SignerName,
			Extra: map[string]capi.ExtraValue{
				"issuer":  []string{issuerAndProfile[0]},
				"profile": []string{issuerAndProfile[1]},
			},
		},
	}

	_, err = client.Get(ctx, certificateSigningRequestName, metaV1.GetOptions{})
	if err != nil {
		logger.KV(xlog.DEBUG, "status", "unable to get CSR", "err", err.Error())
		_, err = client.Create(ctx, certificateSigningRequest, metaV1.CreateOptions{})
		if err != nil {
			return errors.WithMessage(err, "unable to create the certificate signing request")
		}
		logger.KV(xlog.INFO, "status", "waiting for certificate")
	} else {
		logger.KV(xlog.INFO, "status", "signing request already exists")
	}

	var certificate []byte
	for {
		csr, err := client.Get(ctx, certificateSigningRequestName, metaV1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// If the request got deleted, waiting won't help.
				return errors.Errorf("certificate signing request not found: " + certificateSigningRequestName)
			}
			logger.KV(xlog.ERROR, "status", "unable to retrieve certificate signing request",
				"name", certificateSigningRequestName,
				"err", err.Error(),
			)
			time.Sleep(5 * time.Second)
			continue
		}

		certificate = csr.Status.Certificate
		if len(certificate) > 0 {
			logger.KV(xlog.INFO, "status", "got certificate")
			break
		}

		for _, condition := range csr.Status.Conditions {
			if condition.Type == capi.CertificateDenied {
				return errors.Errorf("certificate signing request (%s) denied for %q: %q", certificateSigningRequestName, condition.Reason, condition.Message)
			}
		}

		logger.KV(xlog.INFO, "status", "csr not issued",
			"name", certificateSigningRequestName,
			"retry", "in 5 seconds",
		)

		time.Sleep(5 * time.Second)
	}

	chain, err := certutil.ParseChainFromPEM(certificate)
	if err != nil {
		logger.Errorf("failed to parse chain: %+v", err)
	} else {
		b := new(strings.Builder)
		print.Certificates(b, chain, false)
		logger.KV(xlog.DEBUG, "cert", b.String())
	}

	certFile := path.Join(r.CertDir, "tls.crt")
	if err := ioutil.WriteFile(certFile, certificate, 0644); err != nil {
		return errors.WithMessage(err, "unable to save certificate")
	}

	logger.KV(xlog.INFO, "status", "wrote_cert", "file", certFile)

	return nil
}

func (r *Request) requestName() (name string) {
	name = fmt.Sprintf("%s-%s-", r.PodName, r.Namespace)
	for i := 0; i < 5; i++ {
		n, err := rand.Int(rand.Reader, randAlphabetLength)
		if err != nil {
			logger.Panicf("failed to generate request name: %v", err)
		}
		name += string(randAlphabet[n.Int64()])
	}
	return
}
