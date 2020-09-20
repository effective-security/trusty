package v1

import "time"

// AuthStsURLResponse provides response for AuthStsURLRequest
type AuthStsURLResponse struct {
	URL string `json:"url"`
}

// AuthState is OAuth state provided by an authenticating client
type AuthState struct {
	RedirectURL string `json:"redirect_url"`
	DeviceID    string `json:"device_id"`
}

// UserInfo provides basic info about user
type UserInfo struct {
	ID    string `json:"user_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Authorization is returned to the client in token refresh response
type Authorization struct {
	Version     string    `json:"version"`
	DeviceID    string    `json:"device_id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	UserName    string    `json:"user_name"`
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
