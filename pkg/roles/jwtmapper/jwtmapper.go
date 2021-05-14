package jwtmapper

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "jwtmapper")

// ProviderName is identifier for role mapper provider
const ProviderName = "jwt"

// Key for JWT signature
type Key struct {
	// ID of the key
	ID   string `json:"id" yaml:"id"`
	Seed string `json:"seed" yaml:"seed"`
}

// Config provides OAuth2 configuration
type Config struct {
	// Audience specifies audience claim
	Audience string `json:"audience" yaml:"audience"`
	// Issuer specifies issuer claim
	Issuer string `json:"issuer" yaml:"issuer"`
	// KeyID specifies ID of the current key
	KeyID string `json:"kid" yaml:"kid"`
	// Keys specifies list of issuer's keys
	Keys []*Key `json:"keys" yaml:"keys"`
	// DefaultRole specifies default role name
	DefaultRole string `json:"default_role" yaml:"default_role"`
	// RolesMap is a map of roles to list of users
	RolesMap map[string][]string `json:"roles" yaml:"roles"`
}

// Provider of OAuth2 identity
type Provider struct {
	issuer   string
	audience string
	kid      string
	keys     map[string][]byte
	role     string
	roles    map[string]string
}

// LoadConfig returns configuration loaded from a file
func LoadConfig(file string) (*Config, error) {
	if file == "" {
		return &Config{}, nil
	}

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var config Config
	if strings.HasSuffix(file, ".json") {
		err = json.Unmarshal(raw, &config)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to unmarshal JSON: %q", file)
		}
	} else {
		err = yaml.Unmarshal(raw, &config)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to unmarshal YAML: %q", file)
		}
	}

	if config.KeyID == "" {
		return nil, errors.Errorf("missing kid: %q", file)
	}
	if len(config.Keys) == 0 {
		return nil, errors.Errorf("missing keys: %q", file)
	}

	return &config, nil
}

// Load returns new Provider
func Load(cfgfile string) (*Provider, error) {
	cfg, err := LoadConfig(cfgfile)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return New(cfg), nil
}

// New returns new Provider
func New(cfg *Config) *Provider {
	p := &Provider{
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		kid:      cfg.KeyID,
		keys:     map[string][]byte{},
		role:     cfg.DefaultRole,
		roles:    map[string]string{},
	}

	if p.issuer == "" {
		p.issuer = ProviderName
	}
	if p.audience == "" {
		p.audience = ProviderName
	}
	if p.role == "" {
		p.role = identity.GuestRoleName
	}

	if len(cfg.Keys) > 0 {
		for _, key := range cfg.Keys {
			p.keys[key.ID] = certutil.SHA256([]byte(key.Seed))
		}

		if p.kid == "" {
			p.kid = cfg.Keys[len(cfg.Keys)-1].ID
		}
	}

	for role, users := range cfg.RolesMap {
		for _, user := range users {
			p.roles[user] = role
		}
	}

	return p
}

// CurrentKey returns the key currently being used to sign tokens.
// Made pubic for tests. Used to let the test look inside the otherwise opaque service token.
// Instead, consider a refactor so that mapping a token isn't tied in with the http request handling.
func (p *Provider) CurrentKey() (string, []byte) {
	if key, ok := p.keys[p.kid]; ok {
		return p.kid, key
	}
	for id, key := range p.keys {
		return id, key
	}
	return "", nil
}

func (p *Provider) userRole(claims *TrustyClaims) string {
	role := p.roles[claims.UserInfo.Email]
	if role == "" {
		// if not found, keep using the default
		role = p.role
	}

	return role
}

// SignToken returns signed JWT token with custom claims
func (p *Provider) SignToken(userInfo *v1.UserInfo, deviceID string, expiry time.Duration) (*v1.Authorization, error) {
	kid, key := p.CurrentKey()
	now := time.Now().UTC()
	expiresAt := now.Add(expiry)
	claims := &TrustyClaims{
		userInfo,
		deviceID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    p.issuer,
			IssuedAt:  now.Unix(),
			Audience:  p.audience,
			Subject:   userInfo.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = kid

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(key)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to sign token")
	}

	auth := &v1.Authorization{
		Version:     "v1.0",
		DeviceID:    deviceID,
		UserID:      userInfo.ID,
		Login:       userInfo.Login,
		Email:       userInfo.Email,
		Name:        userInfo.Name,
		Role:        p.userRole(claims),
		TokenType:   "jwt",
		AccessToken: tokenString,
		ExpiresAt:   expiresAt,
		IssuedAt:    now,
	}

	return auth, nil
}

// Applicable returns true if the request has autherization data applicable to the provider
func (p *Provider) Applicable(r *http.Request) bool {
	key := r.Header.Get(header.Authorization)
	return key != "" && strings.HasPrefix(key, header.Bearer)
}

// IdentityMapper interface
func (p *Provider) IdentityMapper(r *http.Request) (identity.Identity, error) {
	parts := strings.Split(r.Header.Get(header.Authorization), " ")
	if len(parts) != 2 || parts[0] != header.Bearer {
		return nil, nil
	}

	deviceID := r.Header.Get(header.XDeviceID)
	claims := &TrustyClaims{
		nil,
		deviceID,
		jwt.StandardClaims{
			Issuer:   p.issuer,
			Audience: p.audience,
		},
	}

	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if kid, ok := token.Header["kid"]; ok {
			var id string
			switch t := kid.(type) {
			case string:
				id = t
			case int:
				id = strconv.Itoa(t)
			}

			if key, ok := p.keys[id]; ok {
				return key, nil
			}
			return nil, errors.Errorf("unexpected kid")
		}
		return nil, errors.Errorf("missing kid")
	})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to verify token")
	}

	if claims, ok := token.Claims.(*TrustyClaims); ok && token.Valid {
		/* TODO: restore after trustyctl passes the header
		if claims.DeviceID != deviceID {
			return nil, errors.Errorf("invalid deviceID: %s", deviceID)
		}
		*/
		if claims.Issuer != p.issuer {
			return nil, errors.Errorf("invalid issuer: %s", claims.Issuer)
		}
		if claims.Audience != p.audience {
			return nil, errors.Errorf("invalid audience: %s", claims.Audience)
		}

		role := p.userRole(claims)

		return identity.NewIdentityWithUserInfo(role, claims.UserInfo.Email, claims.UserInfo.ID, claims.UserInfo), nil
	}

	return nil, errors.Errorf("invalid token")
}

// TrustyClaims for OAuth token
type TrustyClaims struct {
	UserInfo *v1.UserInfo `json:"trusty"` // user info from STS
	DeviceID string       `json:"device_id"`
	jwt.StandardClaims
}
