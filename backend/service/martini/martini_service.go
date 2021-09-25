package martini

import (
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db/cadb"
	"github.com/martinisecurity/trusty/backend/db/orgsdb"
	"github.com/martinisecurity/trusty/pkg/email"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/pkg/payment"
)

// ServiceName provides the Service Name for this package
const ServiceName = "martini"

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/backend/service", "martini")

// Service defines the Status service
type Service struct {
	FccBaseURL string

	disableEmail bool
	server       *gserver.Server
	cfg          *config.Configuration
	db           orgsdb.OrgsDb
	cadb         cadb.CaReadonlyDb
	emailProv    *email.Provider
	paymentProv  payment.Provider
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, db orgsdb.OrgsDb, cadb cadb.CaReadonlyDb, emailProv *email.Provider, paymentProv payment.Provider) error {
		svc := &Service{
			server:      server,
			cfg:         cfg,
			db:          db,
			cadb:        cadb,
			emailProv:   emailProv,
			paymentProv: paymentProv,
		}

		server.AddService(svc)
		return nil
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return ServiceName
}

// IsReady indicates that the service is ready to serve its end-points
func (s *Service) IsReady() bool {
	return true
}

// Close the subservices and it's resources
func (s *Service) Close() {
}

// DisableEmail is used in unit test
// TODO: mock Email provider
func (s *Service) DisableEmail() {
	s.disableEmail = true
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
	r.GET(v1.PathForMartiniSearchCorps, s.SearchCorpsHandler())
	r.GET(v1.PathForMartiniCerts, s.GetCertsHandler())
	r.GET(v1.PathForMartiniOrgMembers, s.GetOrgMembersHandler())
	r.POST(v1.PathForMartiniOrgMembers, s.OrgMemberHandler())
	r.GET(v1.PathForMartiniOrgAPIKeys, s.GetOrgAPIKeysHandler())
	r.POST(v1.PathForMartiniRegisterOrg, s.RegisterOrgHandler())
	r.POST(v1.PathForMartiniApproveOrg, s.ApproveOrgHandler())
	r.POST(v1.PathForMartiniValidateOrg, s.ValidateOrgHandler())
	r.POST(v1.PathForMartiniDeleteOrg, s.DeleteOrgHandler())
	r.GET(v1.PathForMartiniOrgs, s.GetOrgsHandler())
	r.GET(v1.PathForMartiniOrgByID, s.GetOrgHandler())
	r.GET(v1.PathForMartiniSearchOrgs, s.SearchOrgsHandler())

	r.POST(v1.PathForMartiniCreateSubscription, s.CreateSubsciptionHandler())
	r.POST(v1.PathForMartiniCancelSubscription, s.CancelSubsciptionHandler())
	r.GET(v1.PathForMartiniListSubscriptions, s.ListSubsciptionsHandler())
	r.GET(v1.PathForMartiniSubscriptionsProducts, s.SubscriptionsProductsHandler())
	r.POST(v1.PathForMartiniStripeWebhook, s.StripeWebhookHandler())

	r.GET(v1.PathForMartiniFccFrn, s.FccFrnHandler())
	r.GET(v1.PathForMartiniFccContact, s.FccContactHandler())

	// TODO: remove after Web updated
	r.POST("/v1/ms/register_org", s.RegisterOrgHandler())
	r.POST("/v1/ms/approve_org", s.ApproveOrgHandler())
	r.POST("/v1/ms/validate_org", s.ValidateOrgHandler())
	r.POST("/v1/ms/delete_org", s.DeleteOrgHandler())
}

// Db returns DB
// Used in Unittests
func (s *Service) Db() orgsdb.OrgsDb {
	return s.db
}

// CaDb returns CaReadonlyDb
// Used in Unittests
func (s *Service) CaDb() cadb.CaReadonlyDb {
	return s.cadb
}

// PaymentProvider returns paymentProv
// Used in Unittests
func (s *Service) PaymentProvider() payment.Provider {
	return s.paymentProv
}
