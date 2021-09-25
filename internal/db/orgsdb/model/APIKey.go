package model

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
)

// APIKey provides API key
type APIKey struct {
	ID         uint64    `json:"id"`
	OrgID      uint64    `json:"org_id"`
	Key        string    `json:"key"`
	Enrollemnt bool      `json:"enrollment"`
	Management bool      `json:"management"`
	Billing    bool      `json:"billing"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	UsedAt     time.Time `json:"used_at"`
}

// Validate returns error if the model is not valid
func (k *APIKey) Validate() error {
	if k.OrgID == 0 {
		return errors.Errorf("invalid ID")
	}
	if len(k.Key) < 32 || len(k.Key) > 64 {
		return errors.Errorf("invalid key: %q", k.Key)
	}
	return nil
}

// ToDto converts model to v1.APIKey DTO
func (k *APIKey) ToDto() *v1.APIKey {
	return &v1.APIKey{
		ID:         strconv.FormatUint(k.ID, 10),
		OrgID:      strconv.FormatUint(k.OrgID, 10),
		Key:        k.Key,
		Enrollemnt: k.Enrollemnt,
		Management: k.Management,
		Billing:    k.Billing,
		CreatedAt:  k.CreatedAt,
		ExpiresAt:  k.ExpiresAt,
		UsedAt:     k.UsedAt,
	}
}

// ToAPIKeysDto returns API Keys
func ToAPIKeysDto(list []*APIKey) []v1.APIKey {
	res := make([]v1.APIKey, len(list))
	for i, key := range list {
		res[i] = *key.ToDto()
	}
	return res
}

// GenerateAPIKey returns random key
func GenerateAPIKey() string {
	return base64.RawURLEncoding.EncodeToString(certutil.Random(24))
}
