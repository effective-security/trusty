package email

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/mailgun/mailgun-go/v4"
	"gopkg.in/yaml.v2"
)

// MailgunProviderName is mailgun provider name
const MailgunProviderName = "mailgun"

// MailgunClientConfig config for Mailgun
type MailgunClientConfig struct {
	// ProviderID specifies email provider id
	ProviderID string `json:"provider_id" yaml:"provider_id"`
	// Domain specifies domain emails are sent from
	Domain string `json:"domain" yaml:"domain"`
	// Sender specifies sender email
	Sender string `json:"sender" yaml:"sender"`
	// APIKey specifies mailgun private key
	APIKey string `json:"api_key" yaml:"api_key"`
}

// MailgunClient client for Mailgun
type MailgunClient struct {
	cfg *MailgunClientConfig
}

// NewMailgunClient creates a new mailgun client
func NewMailgunClient(cfg *MailgunClientConfig) *MailgunClient {
	return &MailgunClient{
		cfg: cfg,
	}
}

// Send implements email sender for users
func (m *MailgunClient) Send(to string, subject string, htmlBody string) error {
	mg := mailgun.NewMailgun(m.cfg.Domain, m.cfg.APIKey)
	message := mg.NewMessage(m.cfg.Sender, subject, "", to)

	message.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err := mg.Send(ctx, message)

	if err != nil {
		return errors.Annotatef(err, "failed to send email using mailgun provider: domain=%q, sender=%q, to=%q", m.cfg.Domain, m.cfg.Sender, to)
	}

	return nil
}

// LoadMailgunConfig returns mailgun configuration loaded from a file
func LoadMailgunConfig(file string) (*MailgunClientConfig, error) {
	if file == "" {
		return &MailgunClientConfig{}, nil
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var config MailgunClientConfig
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
