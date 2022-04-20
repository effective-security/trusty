package accesstoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/effective-security/xpki/dataprotection"
	"github.com/effective-security/xpki/jwt"
	"github.com/pkg/errors"
)

// Provider of Access Token
type Provider struct {
	dp     dataprotection.Provider
	parser jwt.Parser
}

// New returns new Provider
func New(dp dataprotection.Provider, parser jwt.Parser) *Provider {
	return &Provider{
		dp:     dp,
		parser: parser,
	}
}

// Protect returns AccessToken from claims
func (p *Provider) Protect(ctx context.Context, claims jwt.MapClaims) (string, error) {
	js, err := json.Marshal(claims)
	if err != nil {
		return "", errors.WithStack(err)
	}

	protected, err := p.dp.Protect(ctx, js)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return "pat." + base64.RawURLEncoding.EncodeToString(protected), nil
}

// Claims returns claims from the Access Token,
// or nil if `auth` is not Access Token
func (p *Provider) Claims(ctx context.Context, auth string) (jwt.MapClaims, error) {
	if !strings.HasPrefix(auth, "pat.") {
		if p.parser == nil {
			// not supported
			return nil, nil
		}
		cl, err := p.parser.ParseToken(auth, jwt.VerifyConfig{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return cl, nil
	}

	protected, err := base64.RawURLEncoding.DecodeString(auth[4:])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	js, err := p.dp.Unprotect(ctx, protected)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	claims := jwt.MapClaims{}
	err = json.Unmarshal(js, &claims)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return claims, nil
}
