package model_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKey(t *testing.T) {
	now := time.Now()
	u := &model.APIKey{
		ID:         1000,
		OrgID:      2000,
		Key:        model.GenerateAPIKey(),
		Enrollemnt: true,
		Management: true,
		Billing:    true,
		CreatedAt:  now,
		UsedAt:     now.Add(time.Hour),
		ExpiresAt:  now.Add(2 * time.Hour),
	}
	assert.Equal(t, len(u.Key), 32)

	dto := u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, "2000", dto.OrgID)
	assert.Equal(t, u.Key, dto.Key)
	assert.Equal(t, u.Enrollemnt, dto.Enrollemnt)
	assert.Equal(t, u.Management, dto.Management)
	assert.Equal(t, u.Billing, dto.Billing)
	assert.Equal(t, u.ExpiresAt, dto.ExpiresAt)
	assert.Equal(t, u.CreatedAt, dto.CreatedAt)
	assert.Equal(t, u.UsedAt, dto.UsedAt)

	ul := []*model.APIKey{u}
	assert.Len(t, model.ToAPIKeysDto(ul), 1)
}
func TestAPIKeyValidate(t *testing.T) {
	tcases := []struct {
		u   *model.APIKey
		err string
	}{
		{&model.APIKey{}, "invalid ID"},
		{&model.APIKey{OrgID: 123}, "invalid key: \"\""},
		{&model.APIKey{OrgID: 123, Key: "11782345691872364981723649871263948716239841178234569187236498172364987126394871623984"}, "invalid key: \"11782345691872364981723649871263948716239841178234569187236498172364987126394871623984\""},
		{&model.APIKey{OrgID: 123, Key: "01234567890123456789012345678901"}, ""},
	}
	for _, tc := range tcases {
		err := tc.u.Validate()
		if tc.err != "" {
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGenerateAPIKey(t *testing.T) {
	for i := 0; i < 1000; i++ {
		k := model.GenerateAPIKey()
		_, err := base64.RawURLEncoding.DecodeString(k)
		assert.NoError(t, err, k)
	}

	_, err := base64.RawURLEncoding.DecodeString("wB4XqRPi6MfA2kKay-J3fubc37Z4MVUp")
	assert.NoError(t, err)
}
