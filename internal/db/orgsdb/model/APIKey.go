package model

import (
	"time"

	"github.com/juju/errors"
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
func (t *APIKey) Validate() error {
	if t.OrgID == 0 {
		return errors.Errorf("invalid ID")
	}
	if len(t.Key) != 32 {
		return errors.Errorf("invalid key: %q", t.Key)
	}
	return nil
}
