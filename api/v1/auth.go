package v1

import "time"

// AuthStsURLResponse provides response for AuthStsURLRequest
type AuthStsURLResponse struct {
	URL string `json:"url"`
}

// UserInfo provides basic info about user
type UserInfo struct {
	ID         string `json:"id"`
	ExternalID string `json:"extern_id"`
	Provider   string `json:"provider"`
	Login      string `json:"login"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Company    string `json:"company,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
}

// Authorization is returned to the client in token refresh response
// TODO: add refresh token
type Authorization struct {
	Version     string    `json:"version"`
	DeviceID    string    `json:"device_id"`
	UserID      string    `json:"user_id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	TokenType   string    `json:"token_type"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	IssuedAt    time.Time `json:"issued_at"`
}

// AuthTokenRefreshResponse provides response for token refresh request
type AuthTokenRefreshResponse struct {
	Authorization *Authorization `json:"authorization"`
	Profile       *UserInfo      `json:"profile"`
}
