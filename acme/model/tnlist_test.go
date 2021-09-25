package model_test

import (
	"testing"

	"github.com/martinisecurity/trusty/acme/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTNEntry(t *testing.T) {
	tn, err := model.ParseTNEntry("MAigBhYENzA5Sg==")
	require.NoError(t, err)
	assert.Equal(t, "709J", tn.SPC.Code)

	tn = &model.TNEntry{
		SPC: model.ServiceProviderCode{
			Code: "709J",
		},
	}
	b64, err := tn.Base64()
	require.NoError(t, err)
	assert.Equal(t, "MAigBhYENzA5Sg==", b64)
}
