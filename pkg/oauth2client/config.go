package oauth2client

// Config provides OAuth2 configuration
type Config struct {
	// ProviderID specifies Auth.Provider ID
	ProviderID string `json:"provider_id" yaml:"provider_id"`
	// ClientID specifies client ID
	ClientID string `json:"client_id" yaml:"client_id"`
	// ClientSecret specifies client secret
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	// Scopes specifies the list of scopes
	Scopes []string `json:"scopes" yaml:"scopes"`
	// ResponseType specifies the response type, default is "code"
	ResponseType string `json:"response_type" yaml:"response_type"`
	// AuthURL specifies auth URL
	AuthURL string `json:"auth_url" yaml:"auth_url"`
	// TokenURL specifies token URL
	TokenURL string `json:"token_url"  yaml:"token_url"`
	// PubKey specifies PEM encoded Public Key of the JWT issuer
	PubKey string `json:"pubkey" yaml:"pubkey"`
	// Audience of JWT token
	Audience string `json:"audience" yaml:"audience"`
	// Issuer of JWT token
	Issuer string `json:"issuer" yaml:"issuer"`
}
