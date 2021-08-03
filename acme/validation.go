package acme

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/juju/errors"
)

// UpdateAuthorizationChallenge updates Authorization challenge
func (d *Provider) UpdateAuthorizationChallenge(ctx context.Context, authz *model.Authorization, challIdx int) (*model.Authorization, error) {
	now := time.Now().UTC()
	if authz.ExpiresAt.IsZero() || authz.ExpiresAt.Before(now) {
		return nil, errors.NotFoundf("authorization %s/%s has expired on %q",
			authz.RegistrationID, authz.ID, authz.ExpiresAt.Format(time.RFC3339))
	}

	if challIdx < 0 || challIdx >= len(authz.Challenges) {
		return nil, errors.Errorf("challenge #%d not found", challIdx)
	}

	chall := &authz.Challenges[challIdx]

	if !d.ChallengeTypeEnabled(string(chall.Type), authz.RegistrationID) {
		// TODO: TLSSNIRevalidation, check for existing cert
		return nil, errors.Errorf("challenge type %q no longer allowed", chall.Type)
	}

	if authz.Status == v2acme.StatusValid && d.cfg.Policy.ReuseValidAuthz {
		logger.Infof("reason=reuse_valid, acct=%d, authz=%d",
			authz.RegistrationID, authz.ID)
		return authz, nil
	}

	// Look up the account key for this authorization
	reg, err := d.GetRegistration(ctx, authz.RegistrationID)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if chall.Type == v2acme.IdentifierDNS {
		// Compute the key authorization field based on the registration key
		expectedKeyAuthorization, err := chall.ExpectedKeyAuthorization(reg.Key)
		if err != nil {
			return nil, errors.Trace(err)
		}

		if chall.KeyAuthorization != "" && chall.KeyAuthorization != expectedKeyAuthorization {
			return nil, errors.Errorf("invalid authorization key for challenge: %d/%d/%d",
				authz.RegistrationID, authz.ID, chall.ID)
		}

		// Populate the KeyAuthorization such that the VA can confirm the
		// expected vs actual without needing the registration key.
		chall.KeyAuthorization = expectedKeyAuthorization
		if err := chall.CheckConsistencyForValidation(); err != nil {
			return nil, errors.Trace(err)
		}
	}

	chall.Status = v2acme.StatusProcessing
	authz.Status = v2acme.StatusProcessing

	authz, err = d.UpdatePendingAuthorization(ctx, authz)
	if err != nil {
		return nil, errors.Trace(err)
	}

	go func(authz *model.Authorization, chall *model.Challenge) {
		// make a copy of the challenges slice here for mutation
		challenges := make([]model.Challenge, len(authz.Challenges))
		copy(challenges, authz.Challenges)
		authz.Challenges = challenges

		ctx := context.Background()

		records, err := d.PerformValidation(ctx, authz.RegistrationID, authz.Identifier, chall)
		if err != nil {
			logger.Errorf("domain=%q, authz=%d, err=[%v]",
				authz.Identifier.Value, authz.ID, errors.ErrorStack(err))
		}

		if idx, ok := authz.FindChallenge(chall.ID); ok {
			chall = &authz.Challenges[idx]
		} else {
			// this is unexpected
			logger.Errorf("reason=challenge_not_found, authz=%d, challenge=%d",
				authz.ID, chall.ID)
			return
		}

		chall.ValidationRecord = records

		if err != nil {
			if p := v2acme.IsProblem(err); p != nil {
				// move to "invalid" state if v2acme.Problem is returned
				chall.Status = v2acme.StatusInvalid
				chall.Error = p
			} else if err != nil {
				// for other errors, keep retrying in "pending" state
				chall.Status = v2acme.StatusPending
				chall.Error = &v2acme.Problem{
					Type:   v2acme.ServerInternalProblem,
					Detail: "unable to validate challenge: " + err.Error(),
				}
			}
		} else {
			chall.Status = v2acme.StatusValid
			chall.Error = nil
		}

		_, err = d.UpdateAuthorizationAfterValidation(ctx, authz)
		if err != nil {
			logger.Errorf("reason=UpdateAuthorizationAfterValidation, authz=%d, err=[%v]",
				authz.ID, errors.ErrorStack(err))
		}
	}(authz, chall)

	return authz, nil
}

// PerformValidation validates the challenge
func (d *Provider) PerformValidation(ctx context.Context, regID uint64, idn v2acme.Identifier, chall *model.Challenge) ([]model.ValidationRecord, error) {
	switch idn.Type {
	case v2acme.IdentifierTNAuthList:
		return d.ValidateTNAuthList(ctx, regID, idn.Value, chall)
	}
	return nil, errors.Errorf("unsupported challenge type: %s", idn.Type)
}

// TkClaims for SPC
type TkClaims struct {
	Exp int64  `json:"exp"`
	JTI string `json:"jti"`
	ATC struct {
		TKType      string `json:"tktype"`
		TKValue     string `json:"tkvalue"`
		CA          bool   `json:"ca"`
		Fingerprint string `json:"fingerprint"`
	} `json:"atc"`
	jwt.StandardClaims
}

// ValidateTNAuthList validates TNAuthList
func (d *Provider) ValidateTNAuthList(ctx context.Context, regID uint64, idn string, chall *model.Challenge) ([]model.ValidationRecord, error) {
	var challenge map[string]string
	err := json.Unmarshal([]byte(chall.KeyAuthorization), &challenge)
	if err != nil {
		return nil, errors.Annotate(err, "invalid challenge value")
	}

	atc := challenge["atc"]

	claims := &TkClaims{}

	parsed, err := jwt.ParseWithClaims(atc, claims, func(token *jwt.Token) (interface{}, error) {
		return d.stipaChain[0].PublicKey, nil
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to validate SPC token signature")
	}
	if parsed.Header["alg"].(string) != "ES256" {
		return nil, errors.Annotate(err, "unsupported SPC token alg")
	}
	claims, ok := parsed.Claims.(*TkClaims)
	if !ok || !parsed.Valid {
		return nil, errors.Errorf("invalid SPC token")
	}

	if claims.ATC.TKType != string(v2acme.IdentifierTNAuthList) ||
		claims.ATC.TKValue != idn {
		return nil, errors.Errorf("invalid SPC token TKValue")
	}

	if time.Now().After(time.Unix(claims.Exp, 0)) {
		return nil, errors.Errorf("expired SPC token")
	}

	if regID != 0 {
		// finally check the fingerprint
		reg, err := d.db.GetRegistration(ctx, regID)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to find registration %d", regID)
		}

		fp := strings.TrimLeft(claims.ATC.Fingerprint, "SHA256 ")
		fp = strings.ReplaceAll(fp, ":", "")

		if strings.EqualFold(reg.KeyID, fp) {
			return nil, errors.Errorf("SPC fingerprint does not match")
		}
	}
	return nil, nil
}
