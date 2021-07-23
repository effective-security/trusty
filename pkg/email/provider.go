package email

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/go-phorce/dolly/fileutil"
	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

// Client exposes send email operations
type Client interface {
	Send(to string, subject string, htmlBody string) error
}

// Provider of email clients
type Provider struct {
	clients map[string]Client
}

// Config provides email providers common configuration
type Config struct {
	// ProviderID specifies email provider id
	ProviderID string `json:"provider_id" yaml:"provider_id"`
}

// NewProvider returns provider
func NewProvider(locations []string) (*Provider, error) {
	p := &Provider{
		clients: make(map[string]Client),
	}
	for _, l := range locations {
		cfg, err := LoadConfig(l)
		if err != nil {
			return nil, errors.Trace(err)
		}

		switch cfg.ProviderID {
		case MailgunProviderName:
			mailGunConfig, err := LoadMailgunConfig(l)
			if err != nil {
				return nil, errors.Trace(err)
			}
			mailGunConfig.APIKey, err = fileutil.LoadConfigWithSchema(mailGunConfig.APIKey)
			if err != nil {
				return nil, errors.Trace(err)
			}

			p.clients[cfg.ProviderID] = NewMailgunClient(mailGunConfig)
		default:
			continue
		}

	}

	return p, nil
}

// GetProvider returns email provider by id
func (p *Provider) GetProvider(id string) Client {
	return p.clients[id]
}

// LoadConfig returns configuration loaded from a file
func LoadConfig(file string) (*Config, error) {
	if file == "" {
		return &Config{}, nil
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var config Config
	if strings.HasSuffix(file, ".json") {
		err = json.Unmarshal(b, &config)
	} else {
		err = yaml.Unmarshal(b, &config)
	}
	if err != nil {
		return nil, errors.Annotatef(err, "unable to unmarshal %q", file)
	}

	return &config, nil
}
