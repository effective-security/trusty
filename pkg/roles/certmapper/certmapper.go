package certmapper

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-phorce/dolly/algorithms/slices"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"gopkg.in/yaml.v2"
)

// ProviderName is identifier for role mapper provider
const ProviderName = "cert"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "certmapper")

// Identity of the caller
type Identity struct {
	// Name of identity
	Name string `json:"name" yaml:"name"`
	// Role of identity
	Role string `json:"role" yaml:"role"`
}

// Config provides mapping of Subject Names to Roles
type Config struct {
	// NamesMap is a map of role to X509 Subjects
	NamesMap map[string][]string `json:"roles" yaml:"roles"`
	// ValidOrganizations is a list of accepted Organization values from a cert.
	ValidOrganizations []string `json:"valid_organizations" yaml:"valid_organizations"`
	// ValidIssuers is a list of accepted root Subject names
	ValidIssuers []string `json:"valid_issuers" yaml:"valid_issuers"`
}

// Provider of Cert identity
type Provider struct {
	namesMap      map[*regexp.Regexp]identity.Identity
	organizations []string
	// list of accepted root Subject names
	issuers []string
}

// LoadConfig returns configuration loaded from a file
func LoadConfig(file string) (*Config, error) {
	if file == "" {
		return &Config{}, nil
	}

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var config Config

	if strings.HasSuffix(file, ".json") {
		err = json.Unmarshal(raw, &config)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to unmarshal JSON: %q", file)
		}
	} else {
		err = yaml.Unmarshal(raw, &config)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to unmarshal YAML: %q", file)
		}
	}
	return &config, nil
}

// Load returns new Provider
func Load(cfgfile string) (*Provider, error) {
	cfg, err := LoadConfig(cfgfile)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return New(cfg)
}

// New returns new Provider
func New(cfg *Config) (*Provider, error) {
	p := &Provider{
		namesMap:      map[*regexp.Regexp]identity.Identity{},
		organizations: cfg.ValidOrganizations,
		issuers:       cfg.ValidIssuers,
	}

	for role, subjects := range cfg.NamesMap {
		for _, subj := range subjects {
			subj = strings.ToLower(subj)
			// note that the order is important, first replace dots than stars
			subjReplaced := strings.Replace(subj, ".", "[.]", -1)
			subjReplaced = strings.Replace(subjReplaced, "*", ".*", -1)
			subjReplaced = "^" + subjReplaced + "$"
			r, err := regexp.Compile(subjReplaced)
			if err != nil {
				return nil, errors.Errorf("invalid regex %q, subj=%q", subjReplaced, subj)
			}
			p.namesMap[r] = subjectToIdentity(role, subj)
		}
	}
	return p, nil
}

// Applicable returns true if the request has autherization data applicable to the provider
func (p *Provider) Applicable(r *http.Request) bool {
	return r.TLS != nil && len(r.TLS.PeerCertificates) > 0
}

// IdentityMapper interface
func (p *Provider) IdentityMapper(r *http.Request) (identity.Identity, error) {
	if !p.Applicable(r) {
		return nil, nil
	}

	return p.identity(r.TLS)
}

// ApplicableForContext returns true if the provider is applicable for context
func (p *Provider) ApplicableForContext(ctx context.Context) bool {
	c, ok := peer.FromContext(ctx)
	if ok {
		si, ok := c.AuthInfo.(credentials.TLSInfo)
		return ok && len(si.State.PeerCertificates) > 0
	}

	return false
}

// IdentityFromContext returns identity from context
func (p *Provider) IdentityFromContext(ctx context.Context) (identity.Identity, error) {
	c, ok := peer.FromContext(ctx)
	if ok {
		si, ok := c.AuthInfo.(credentials.TLSInfo)
		if ok {
			return p.identity(&si.State)
		}
	}

	return nil, nil
}

func (p *Provider) identity(TLS *tls.ConnectionState) (identity.Identity, error) {
	var id identity.Identity
	var org, issuer string

	peers := TLS.PeerCertificates
	if len(p.organizations) > 0 {
		found := false
		for _, peer := range peers {
			for _, org = range peer.Subject.Organization {
				if found = slices.ContainsString(p.organizations, org); found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, errors.Errorf("the %q organization is not allowed", peers[0].Subject.Organization[0])
		}
	}
	if len(p.issuers) > 0 {
		found := false
		for _, chain := range TLS.VerifiedChains {
			issuer = certutil.NameToString(&chain[len(chain)-1].Subject)
			if found = slices.ContainsString(p.issuers, issuer); found {
				break
			}
		}
		if !found {
			return nil, errors.Errorf("the %q root CA is not allowed", issuer)
		}
	}

	peer := peers[0]
	subj := certutil.NameToString(&peer.Subject)
	subj = strings.ToLower(subj)

	allowed, id := p.isSubjectAllowed(subj)
	if allowed {
		if id != nil {
			logger.Infof("api=IdentityMapper, subject=%q, role=%s, name=%q", subj, id.Role(), id.Name())
			return id, nil
		}
		return nil, errors.Errorf("identity is nil: %q", subj)
	}

	return nil, errors.Errorf("could not determine identity: %q", subj)
}

func (p *Provider) isSubjectAllowed(subject string) (bool, identity.Identity) {
	for regName, identity := range p.namesMap {
		match := regName.MatchString(subject)
		if match {
			return true, identity
		}
	}
	return false, nil
}

func subjectToIdentity(role, subject string) identity.Identity {
	var name string
	for _, token := range strings.Split(subject, ",") {
		token = strings.TrimSpace(token)
		if strings.HasPrefix(token, "cn=") {
			name = token[3:]
			break
		}
	}

	return identity.NewIdentity(role, name, "")
}
