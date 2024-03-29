package trustymain

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/effective-security/porto/pkg/tlsconfig"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/x/fileutil"
	"github.com/effective-security/x/netutil"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/cryptoprov/inmemcrypto"
	"github.com/effective-security/xpki/csr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (a *App) genCert() error {
	if caSrvCfg := a.cfg.HTTPServers[config.CAServerName]; caSrvCfg == nil || caSrvCfg.Disabled {
		return nil
	}

	var ca *authority.Authority
	var db cadb.CaDb
	err := a.container.Invoke(func(a *authority.Authority, d cadb.CaDb) {
		ca = a
		db = d
	})
	if err != nil {
		return errors.Wrap(err, "failed to create authority")
	}
	ctx := context.Background()
	crypto := inmemcrypto.NewProvider()
	for _, gcCfg := range a.cfg.RegistrationAuthority.GenCerts.Profiles {
		if gcCfg.Disabled {
			continue
		}
		ca, err := ca.GetIssuerByProfile(gcCfg.Profile)
		if err != nil {
			return err
		}

		certDir := filepath.Dir(gcCfg.CertFile)
		_ = os.MkdirAll(certDir, 0755)

		keyDir := filepath.Dir(gcCfg.KeyFile)
		if keyDir != certDir {
			_ = os.MkdirAll(keyDir, 0755)
		}

		if fileutil.FileExists(gcCfg.CertFile) == nil &&
			fileutil.FileExists(gcCfg.KeyFile) == nil {
			d, err := time.ParseDuration(gcCfg.Renewal)
			if err != nil {
				logger.KV(xlog.ERROR, "reason", "config", "profile", gcCfg.Profile, "err", err)
				d = 24 * time.Hour
			}
			cutoff := time.Now().Add(d).UTC()
			cert, err := certutil.LoadFromPEM(gcCfg.CertFile)
			if err != nil {
				logger.KV(xlog.INFO, "reason", "load", "profile", gcCfg.Profile, "err", err.Error())
			} else if cutoff.Before(cert.NotAfter.UTC()) {
				// try to load file with the key to ensure it matches and valid
				_, err = tlsconfig.LoadX509KeyPairWithOCSP(gcCfg.CertFile, gcCfg.KeyFile)
				if err == nil {
					logger.KV(xlog.INFO, "reason", "valid", "profile", gcCfg.Profile, "notAfter", cert.NotAfter.Format(time.RFC3339))
					continue
				}
			}
		}

		// Load CSR
		csrf, err := os.ReadFile(gcCfg.CsrProfile)
		if err != nil {
			return errors.WithMessage(err, "read CSR profile")
		}

		req := &csr.CertificateRequest{
			KeyRequest: csr.NewKeyRequest(crypto, gcCfg.Profile, "ecdsa", 256, csr.SigningKey),
		}

		err = yaml.Unmarshal(csrf, &req)
		if err != nil {
			return errors.WithMessagef(err, "invalid CSR profile: %s", gcCfg.CsrProfile)
		}

		if strings.Contains(gcCfg.CsrProfile, "server") ||
			strings.Contains(gcCfg.CsrProfile, "peer") {
			hn, err := os.Hostname()
			if err != nil {
				return errors.WithStack(err)
			}
			req.AddSAN(hn)
			req.AddSAN("localhost")

			ipaddr, err := netutil.WaitForNetwork(1 * time.Second)
			if err == nil && ipaddr != "" {
				req.AddSAN(ipaddr)
			}
		}

		// NOTE: the client cert may have URI with spiffe://
		for _, san := range gcCfg.SAN {
			req.AddSAN(san)
		}

		crt, certPEM, err := ca.GenCert(crypto, req, gcCfg.Profile, gcCfg.CertFile, gcCfg.KeyFile)
		if err != nil {
			return errors.WithStack(err)
		}

		mcert := model.NewCertificate(crt, 0, gcCfg.Profile, string(certPEM), ca.PEM(), a.cfg.ServiceName, nil,
			map[string]string{
				"profile": path.Base(gcCfg.CsrProfile),
				"service": a.cfg.ServiceName,
				"cluster": a.cfg.ClusterName,
				"env":     a.cfg.Environment,
			})
		_, err = db.RegisterCertificate(ctx, mcert)
		if err != nil {
			logger.KV(xlog.ERROR,
				"status", "failed to register certificate",
				"err", err)
			// DO NOT fail on register error
		}
		logger.KV(xlog.NOTICE,
			"status", "signed",
			"cert", gcCfg.CertFile,
			"key", gcCfg.KeyFile,
			"profile", gcCfg.Profile,
			"ikid", mcert.IKID,
			"sn", mcert.SerialNumber,
			"ocsp", crt.OCSPServer,
			"crl", crt.CRLDistributionPoints,
		)
	}
	return nil
}
