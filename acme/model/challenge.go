package model

import (
	"github.com/ekspand/trusty/api/v2acme"
)

// NewChallenge returns an instance of the challenge
func NewChallenge(authzID string, typ v2acme.IdentifierType) *Challenge {
	return &Challenge{
		ID:              GenerateID(),
		AuthorizationID: authzID,
		Type:            typ,
		Status:          v2acme.StatusPending,
		Token:           GenerateToken(),
	}
}
