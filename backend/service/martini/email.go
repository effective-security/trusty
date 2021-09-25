package martini

import (
	"bytes"
	"text/template"

	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/pkg/email"
)

// renderEmailTemplate used to render email template with given data
func renderEmailTemplate(html string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(html)
	if err != nil {
		return "", errors.Annotatef(err, "failed to parse template")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", errors.Annotatef(err, "failed to render email template")
	}

	return buf.String(), nil
}

// sendEmail used to send email with Mailgun provider
func (s *Service) sendEmail(toEmail, subject, htmlTemplate string, data interface{}) error {
	body, err := renderEmailTemplate(htmlTemplate, data)
	if err != nil {
		return errors.Annotatef(err, "failed to render email template")
	}

	if !s.disableEmail {
		logger.KV(xlog.NOTICE, "email", toEmail, subject, subject)

		err = s.emailProv.GetProvider(email.MailgunProviderName).Send(toEmail, subject, body)
		if err != nil {
			return errors.Annotatef(err, "failed to send email to %q", toEmail)
		}
	}
	return nil
}
