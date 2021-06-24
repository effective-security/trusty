package ra

import (
	"context"
	"encoding/hex"
	"sync"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed/proxy"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "ra"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "ra")

// Service defines the Status service
type Service struct {
	server        *gserver.Server
	db            db.Provider
	clientFactory client.Factory
	grpClient     *client.Client
	ca            client.CAClient
	registered    bool
	cfg           *config.Configuration
	lock          sync.RWMutex
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, db db.Provider, clientFactory client.Factory) {
		svc := &Service{
			server:        server,
			cfg:           cfg,
			db:            db,
			clientFactory: clientFactory,
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
	if s.grpClient != nil {
		s.grpClient.Close()
	}
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterRAServiceServer(r, s)
}

// OnStarted is called when the server started and
// is ready to serve requests
func (s *Service) OnStarted() error {
	err := s.registerRoots(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	_, err = s.getCAClient()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *Service) registerRoots(ctx context.Context) error {
	registerCert := func(trust pb.Trust, location string) error {
		crt, err := certutil.LoadFromPEM(location)
		if err != nil {
			return err
		}
		pem, err := certutil.EncodeToPEMString(true, crt)
		if err != nil {
			return err
		}
		c := &model.RootCertificate{
			SKID:             hex.EncodeToString(crt.SubjectKeyId),
			NotBefore:        crt.NotBefore.UTC(),
			NotAfter:         crt.NotAfter.UTC(),
			Subject:          crt.Subject.String(),
			ThumbprintSha256: certutil.SHA256Hex(crt.Raw),
			Trust:            int(trust),
			Pem:              pem,
		}
		_, err = s.db.RegisterRootCertificate(ctx, c)
		if err != nil {
			return err
		}
		logger.Infof("trust=%v, subject=%q", trust, c.Subject)
		return nil
	}

	for _, r := range s.cfg.RegistrationAuthority.PrivateRoots {
		err := registerCert(pb.Trust_Private, r)
		if err != nil {
			logger.Errorf("err=[%v]", errors.ErrorStack(err))
			return err
		}
	}
	for _, r := range s.cfg.RegistrationAuthority.PublicRoots {
		err := registerCert(pb.Trust_Public, r)
		if err != nil {
			logger.Errorf("err=[%v]", errors.ErrorStack(err))
			return err
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.registered = true

	return nil
}

func (s *Service) getCAClient() (client.CAClient, error) {
	var ca client.CAClient
	s.lock.RLock()
	ca = s.ca
	s.lock.RUnlock()
	if ca != nil {
		return ca, nil
	}

	var pb pb.CAServiceServer
	err := s.server.Discovery().Find(&pb)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.ca = client.NewCAClientFromProxy(proxy.CAServerToClient(pb))
		return s.ca, nil
	}

	grpClient, err := s.clientFactory.NewClient("ca")
	if err != nil {
		logger.Errorf("err=[%v]", errors.Details(err))
		return nil, errors.Trace(err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.grpClient != nil {
		s.grpClient.Close()
	}
	s.grpClient = grpClient
	s.ca = grpClient.CAClient()
	return s.ca, nil
}
