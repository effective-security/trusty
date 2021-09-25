package email_test

import (
	"testing"

	"github.com/martinisecurity/trusty/pkg/email"
	"github.com/stretchr/testify/require"
)

func Test_LoadMailgunConfig(t *testing.T) {
	mcfg, err := email.LoadMailgunConfig("testdata/mailgun.yaml")
	require.NoError(t, err)
	require.Equal(t, "mailgun", mcfg.ProviderID)
	require.Equal(t, "martinisecurity.com", mcfg.Domain)
	require.Equal(t, "no-reply@martinisecurity.com", mcfg.Sender)
	require.Equal(t, "env://TRUSTY_MAILGUN_PRIVATE_KEY", mcfg.APIKey)
}
