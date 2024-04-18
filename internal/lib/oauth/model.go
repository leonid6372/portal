package oauth

import (
	"time"
)

// TokenResponse is the authorization server response
type TokenResponse struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // secs
}

// Token structure generated by the authorization server
type Token struct {
	ID           string         `json:"id_token"`
	CreationDate time.Time      `json:"date"`
	ExpiresIn    time.Duration  `json:"expires_in"` // secs
	Credential   string         `json:"credential"`
	Scope        string         `json:"scope"`
	Claims       map[string]int `json:"claims"`
}

// RefreshToken structure included in the authorization server response
type RefreshToken struct {
	CreationDate   time.Time `json:"date"`
	TokenID        string    `json:"id_token"`
	RefreshTokenID string    `json:"id_refresh_token"`
	Credential     string    `json:"credential"`
	Scope          string    `json:"scope"`
}
