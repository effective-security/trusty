package config

// Authz contains configuration for the authorization module
type Authz struct {
	// Allow will allow the specified roles access to this path and its children, in format: ${path}:${role},${role}
	Allow []string `json:"allow" yaml:"allow"`

	// AllowAny will allow any non-authenticated request access to this path and its children
	AllowAny []string `json:"allow_any" yaml:"allow_any"`

	// AllowAnyRole will allow any authenticated request that includes a non empty role
	AllowAnyRole []string `json:"allow_any_role" yaml:"allow_any_role"`

	// LogAllowedAny specifies to log allowed access to Any role
	LogAllowedAny bool `json:"log_allowed_any" yaml:"log_allowed_any"`

	// LogAllowed specifies to log allowed access
	LogAllowed bool `json:"log_allowed" yaml:"log_allowed"`

	// LogDenied specifies to log denied access
	LogDenied bool `json:"log_denied" yaml:"log_denied"`
}

// IdentityMap contains configuration for the roles
type IdentityMap struct {
	// TLS identity map
	TLS TLSIdentityMap `json:"tls" yaml:"tls"`
	// JWT identity map
	JWT JWTIdentityMap `json:"jwt" yaml:"jwt"`
}

// TLSIdentityMap provides roles for TLS
type TLSIdentityMap struct {
	// DefaultAuthenticatedRole specifies role name for identity, if not found in maps
	DefaultAuthenticatedRole string `json:"default_authenticated_role" yaml:"default_authenticated_role"`
	// Enable TLS identities
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Roles is a map of role to TLS identity
	Roles map[string][]string `json:"roles" yaml:"roles"`
}

// JWTIdentityMap provides roles for JWT
type JWTIdentityMap struct {
	// DefaultAuthenticatedRole specifies role name for identity, if not found in maps
	DefaultAuthenticatedRole string `json:"default_authenticated_role" yaml:"default_authenticated_role"`
	// Enable TLS identities
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Audience specifies the token audience
	Audience string `json:"audience" yaml:"audience"`
	// Roles is a map of role to JWT identity
	Roles map[string][]string `json:"roles" yaml:"roles"`
}
