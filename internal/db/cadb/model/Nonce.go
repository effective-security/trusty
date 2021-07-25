package model

import (
	"time"

	"github.com/juju/errors"
)

// Nonce provides nonce
type Nonce struct {
	ID        uint64    `json:"id"`
	Nonce     string    `json:"token"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UsedAt    time.Time `json:"used_at"`
}

// Validate returns error if the model is not valid
func (t *Nonce) Validate() error {
	if len(t.Nonce) != 16 {
		return errors.Errorf("invalid nonce: %q", t.Nonce)
	}
	return nil
}
