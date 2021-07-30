package acme_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ekspand/trusty/acme"
	"github.com/ekspand/trusty/acme/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewController(t *testing.T) {
	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)

	c, err := acme.NewProvider(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, cfg, c.Config())
}

func TestValidateSPC(t *testing.T) {
	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)

	c, err := acme.NewProvider(cfg, nil)
	require.NoError(t, err)

	m := map[string]string{
		"atc": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsIng1dSI6Imh0dHBzOi8vYXV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw",
	}

	js, err := json.Marshal(m)
	require.NoError(t, err)

	_, err = c.ValidateTNAuthList(context.Background(), 0, "MAigBhYENzA5Sg==", &model.Challenge{
		KeyAuthorization: string(js),
	})
	require.NoError(t, err)
}
