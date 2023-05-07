package ca

import (
	"context"
	"sync"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/porto/x/fileutil"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/pkg/certpublisher"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/certutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "ca"

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/service", "ca")

// Service defines the Status service
type Service struct {
	server     *gserver.Server
	ca         *authority.Authority
	db         cadb.CaDb
	publisher  certpublisher.Publisher
	scheduler  tasks.Scheduler
	cfg        *config.Configuration
	registered bool
	lock       sync.RWMutex
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, ca *authority.Authority, db cadb.CaDb, scheduler tasks.Scheduler, publisher certpublisher.Publisher) {
		svc := &Service{
			cfg:       cfg,
			server:    server,
			ca:        ca,
			db:        db,
			publisher: publisher,
			scheduler: scheduler,
		}

		server.AddService(svc)
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return ServiceName
}

// IsReady indicates that the service is ready to serve its end-points
func (s *Service) IsReady() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.registered
}

// Close the subservices and it's resources
func (s *Service) Close() {
	logger.KV(xlog.INFO, "closed", ServiceName)
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r restserver.Router) {
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterCAServer(r, s)
}

// OnStarted is called when the server started and
// is ready to serve requests
func (s *Service) OnStarted() error {
	ctx := context.Background()
	err := s.registerIssuers(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.registerRoots(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	s.registerPublisherTask(ctx)
	return nil
}

// CaDb returns DB
// Used in Unittests
func (s *Service) CaDb() cadb.CaDb {
	return s.db
}

// CA returns Authority
// Used in Unittests
func (s *Service) CA() *authority.Authority {
	return s.ca
}

func (s *Service) registerIssuers(ctx context.Context) error {
	for _, ca := range s.ca.Issuers() {
		bundle := ca.Bundle()
		mcert := model.NewCertificate(bundle.Cert, 0, "ca", bundle.CertPEM, bundle.CACertsPEM, ca.Label(), nil, nil)

		_, err := s.db.RegisterCertificate(ctx, mcert)
		if err != nil {
			logger.KV(xlog.ERROR,
				"status", "failed to register issuer",
				"serial", mcert.SerialNumber,
				"err", err.Error())
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Service) registerCert(ctx context.Context, trust pb.Trust, location string) error {
	crt, err := certutil.LoadFromPEM(location)
	if err != nil {
		return err
	}
	pem, err := certutil.EncodeToPEMString(false, crt)
	if err != nil {
		return err
	}
	c := model.NewRootCertificate(crt, int(trust), pem)
	_, err = s.db.RegisterRootCertificate(ctx, c)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.ContextKV(ctx, xlog.INFO, "trust", trust, "subject", c.Subject)
	return nil
}

func (s *Service) registerRoots(ctx context.Context) error {
	for _, r := range s.cfg.RegistrationAuthority.PrivateRoots {
		if err := fileutil.FileExists(r); err != nil {
			logger.ContextKV(ctx, xlog.WARNING, "err", err.Error())
			continue
		}
		err := s.registerCert(ctx, pb.Trust_Private, r)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR, "err", err)
			return err
		}
	}
	for _, r := range s.cfg.RegistrationAuthority.PublicRoots {
		if err := fileutil.FileExists(r); err != nil {
			logger.ContextKV(ctx, xlog.WARNING, "err", err.Error())
			continue
		}
		err := s.registerCert(ctx, pb.Trust_Public, r)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR, "err", err)
			return err
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.registered = true

	return nil
}

func (s *Service) registerPublisherTask(ctx context.Context) {
	issuers := s.ca.Issuers()
	for _, issuer := range issuers {
		issuer := issuer
		if issuer.CrlRenewal() > 0 && issuer.CrlURL() != "" {
			logger.ContextKV(ctx, xlog.NOTICE,
				"ikid", issuer.SubjectKID(),
				"scheduled", "crl_publisher",
				"interval", issuer.CrlRenewal().String(),
				"crldp", issuer.CrlURL(),
			)

			task := tasks.NewTaskAtIntervals(uint64(issuer.CrlRenewal().Hours()), tasks.Hours)
			taskName := "crl_publisher_" + issuer.SubjectKID()
			task = task.Do(taskName, func() {
				_, err := s.publishCrl(ctx, issuer.SubjectKID())
				if err != nil {
					logger.ContextKV(ctx, xlog.ERROR,
						"ikid", issuer.SubjectKID(),
						"task", taskName,
						"err", err.Error(),
					)
				}
			})
			s.scheduler = s.scheduler.Add(task)
		} else {
			logger.ContextKV(ctx, xlog.NOTICE,
				"ikid", issuer.SubjectKID(),
				"skipped", "crl_publisher",
				"interval", issuer.CrlRenewal().String(),
				"crldp", issuer.CrlURL(),
			)
		}
	}
}
