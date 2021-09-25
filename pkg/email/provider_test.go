package email_test

import (
	"testing"

	"github.com/martinisecurity/trusty/pkg/email"
	"github.com/stretchr/testify/require"
)

func Test_LoadConfig(t *testing.T) {
	cfg, err := email.LoadConfig("testdata/mailgun.yaml")
	require.NoError(t, err)
	require.Equal(t, "mailgun", cfg.ProviderID)
}

func Test_NewProvider(t *testing.T) {
	p, err := email.NewProvider([]string{"testdata/mailgun.yaml"})
	require.NoError(t, err)
	mc := p.GetProvider("mailgun")
	require.NotNil(t, mc)
}
