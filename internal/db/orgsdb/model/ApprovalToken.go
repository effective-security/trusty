package model

import (
	"time"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/internal/db"
)

// ApprovalToken provides approval token
type ApprovalToken struct {
	ID            uint64    `json:"id"`
	OrgID         uint64    `json:"org_id"`
	RequestorID   uint64    `json:"requestor_id"`
	ApproverEmail string    `json:"approver_email"`
	Token         string    `json:"token"`
	Code          string    `json:"code"`
	Used          bool      `json:"used"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	UsedAt        time.Time `json:"used_at"`
}

// Validate returns error if the model is not valid
func (t *ApprovalToken) Validate() error {
	if t.OrgID == 0 || t.RequestorID == 0 {
		return errors.Errorf("invalid ID")
	}
	if t.ApproverEmail == "" || len(t.ApproverEmail) > db.MaxLenForEmail {
		return errors.Errorf("invalid email: %q", t.ApproverEmail)
	}
	if t.Token == "" || len(t.Token) > 16 {
		return errors.Errorf("invalid token: %q", t.Token)
	}
	if t.Code == "" || len(t.Code) > 6 {
		return errors.Errorf("invalid code: %q", t.Code)
	}
	return nil
}
