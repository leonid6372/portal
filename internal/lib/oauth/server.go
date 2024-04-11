package oauth

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"time"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/render"
	"github.com/gofrs/uuid"
)

type GrantType string

const (
	PasswordGrant          GrantType = "password"
	ClientCredentialsGrant GrantType = "client_credentials"
	AuthCodeGrant          GrantType = "authorization_code"
	RefreshTokenGrant      GrantType = "refresh_token"
)

// CredentialsVerifier defines the interface of the user and client credentials verifier.
type CredentialsVerifier interface {
	// Validate username and password returning an error if the user credentials are wrong
	ValidateUser(username, password, scope string, r *http.Request) error
	// Provide additional claims to the token
	AddClaims(credential, tokenID, scope string, r *http.Request) (map[string]int, error)
	// Optionally validate previously stored tokenID during refresh request
	ValidateTokenID(credential, tokenID, refreshTokenID string) error
	// Optionally store the tokenID generated for the user
	StoreTokenID(credential, tokenID, refreshTokenID string) error
}

// BearerServer is the OAuth 2 bearer server implementation.
type BearerServer struct {
	secretKey string
	TokenTTL  time.Duration
	verifier  CredentialsVerifier
	provider  *TokenProvider
}

// NewBearerServer creates new OAuth 2 bearer server
func NewBearerServer(secretKey string, ttl time.Duration, verifier CredentialsVerifier, formatter TokenSecureFormatter) *BearerServer {
	if formatter == nil {
		formatter = NewSHA256RC4TokenSecurityProvider([]byte(secretKey))
	}
	return &BearerServer{
		secretKey: secretKey,
		TokenTTL:  ttl,
		verifier:  verifier,
		provider:  NewTokenProvider(formatter)}
}

// UserCredentials manages password grant type requests
func (bs *BearerServer) UserCredentials(w http.ResponseWriter, r *http.Request) {
	grantType := "password"

	type UserData struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	var userData UserData
	// Декодируем json запроса
	err := render.DecodeJSON(r.Body, &userData)
	// Такую ошибку встретим, если получили запрос с пустым телом.
	// Обработаем её отдельно
	if errors.Is(err, io.EOF) {
		w.WriteHeader(400)
		render.JSON(w, r, resp.Error("empty request"))
		return
	}
	if err != nil {
		w.WriteHeader(400)
		render.JSON(w, r, resp.Error("failed to decode request: "+err.Error()))
		return
	}

	scope := r.FormValue("scope")

	refreshToken := ""
	response, statusCode := bs.generateTokenResponse(GrantType(grantType), userData.Username, userData.Password, refreshToken, scope, "", "", r)

	if statusCode != 200 {
		w.WriteHeader(statusCode)
		render.JSON(w, r, resp.Error(response))
		return
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:  "access_token",
			Value: reflect.Indirect(reflect.ValueOf(response)).FieldByName("Token").String(),
		})

	http.SetCookie(w,
		&http.Cookie{
			Name:  "refresh_token",
			Value: reflect.Indirect(reflect.ValueOf(response)).FieldByName("RefreshToken").String(),
		})

	render.JSON(w, r, resp.OK())
}

// Generate token response
func (bs *BearerServer) generateTokenResponse(grantType GrantType, credential string, secret string, refreshToken string, scope string, code string, redirectURI string, r *http.Request) (interface{}, int) {
	var response *TokenResponse
	switch grantType {
	case PasswordGrant:
		if err := bs.verifier.ValidateUser(credential, secret, scope, r); err != nil {
			return "Not authorized", http.StatusUnauthorized
		}

		token, refresh, err := bs.generateTokens(credential, scope, r)
		if err != nil {
			return "Token generation failed, check claims", http.StatusInternalServerError
		}

		if err = bs.verifier.StoreTokenID(credential, token.ID, refresh.RefreshTokenID); err != nil {
			return "Storing Token ID failed", http.StatusInternalServerError
		}

		if response, err = bs.cryptTokens(token, refresh, r); err != nil {
			return "Token generation failed, check security provider", http.StatusInternalServerError
		}
	case RefreshTokenGrant:
		refresh, err := bs.provider.DecryptRefreshTokens(refreshToken)
		if err != nil {
			return "Not authorized", http.StatusUnauthorized
		}

		if err = bs.verifier.ValidateTokenID(refresh.Credential, refresh.TokenID, refresh.RefreshTokenID); err != nil {
			return "Not authorized invalid token", http.StatusUnauthorized
		}

		token, refresh, err := bs.generateTokens(refresh.Credential, refresh.Scope, r)
		if err != nil {
			return "Token generation failed", http.StatusInternalServerError
		}

		err = bs.verifier.StoreTokenID(refresh.Credential, token.ID, refresh.RefreshTokenID)
		if err != nil {
			return "Storing Token ID failed", http.StatusInternalServerError
		}

		if response, err = bs.cryptTokens(token, refresh, r); err != nil {
			return "Token generation failed", http.StatusInternalServerError
		}
	default:
		return "Invalid grant_type", http.StatusBadRequest
	}

	return response, http.StatusOK
}

func (bs *BearerServer) generateTokens(username, scope string, r *http.Request) (*Token, *RefreshToken, error) {
	token := &Token{ID: uuid.Must(uuid.NewV4()).String(), Credential: username, ExpiresIn: bs.TokenTTL, CreationDate: time.Now().UTC(), Scope: scope}
	if bs.verifier != nil {
		claims, err := bs.verifier.AddClaims(username, token.ID, token.Scope, r)
		if err != nil {
			return nil, nil, err
		}
		token.Claims = claims
	}

	refreshToken := &RefreshToken{RefreshTokenID: uuid.Must(uuid.NewV4()).String(), TokenID: token.ID, CreationDate: time.Now().UTC(), Credential: username, Scope: scope}

	return token, refreshToken, nil
}

func (bs *BearerServer) cryptTokens(token *Token, refresh *RefreshToken, r *http.Request) (*TokenResponse, error) {
	cToken, err := bs.provider.CryptToken(token)
	if err != nil {
		return nil, err
	}
	cRefreshToken, err := bs.provider.CryptRefreshToken(refresh)
	if err != nil {
		return nil, err
	}

	tokenResponse := &TokenResponse{Token: cToken, RefreshToken: cRefreshToken, ExpiresIn: (int64)(bs.TokenTTL / time.Second)}

	return tokenResponse, nil
}
