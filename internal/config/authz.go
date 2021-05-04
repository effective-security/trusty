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

// Identity contains configuration for the Identity mappers
type Identity struct {
	// CertMapper specifies location of the config file for certificate based identity.
	CertMapper string `json:"cert_mapper,omitempty" yaml:"cert_mapper,omitempty"`

	// APIKeyMapper specifies location of the config file for API-Key based identity.
	APIKeyMapper string `json:"api_key_mapper,omitempty" yaml:"api_key_mapper,omitempty"`

	// JWTMapper specifies location of the config file for JWT based identity.
	JWTMapper string `json:"jwt_mapper,omitempty" yaml:"jwt_mapper,omitempty"`
}
