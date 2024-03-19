package auth

import (
	"errors"
	"net/http"
	"os"
	"portal/internal/Storage/postgres/entities/user"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/oauth"
)

var bearerServer *oauth.BearerServer

func GetBearerServer() *oauth.BearerServer {
	return bearerServer
}

func InitBearerServer(log *slog.Logger, storage *postgres.Storage, tokenTTL time.Duration) error {
	const op = "auth.NewBearerServer"
	log.With(slog.String("op", op))

	secret, err := os.ReadFile("C:/Users/Leonid/Desktop/portal/internal/lib/auth/secret.txt")
	if err != nil {
		log.Error("failed to read secret key", sl.Err(err))
		return err
	}

	bearerServer = oauth.NewBearerServer(
		string(secret),
		tokenTTL,
		&UserVerifier{Storage: storage, Log: log},
		nil)

	return nil
}

func GetAuthHandler(log *slog.Logger) func(next http.Handler) http.Handler {
	const op = "auth.GetAuthHandler"
	log.With(slog.String("op", op))

	secret, err := os.ReadFile("C:/Users/Leonid/Desktop/portal/internal/lib/auth/secret.txt")
	if err != nil {
		log.Error("failed to read secret key", sl.Err(err))
	}

	return oauth.Authorize(string(secret), nil)
}

// UserVerifier provides user credentials verifier for testing. Все методы этой структуры нужны для удовлетворения условиям NewBearerServer
type UserVerifier struct {
	Storage *postgres.Storage
	Log     *slog.Logger
}

// ValidateUser validates username and password returning an error if the user credentials are wrong
func (uv *UserVerifier) ValidateUser(username, password, scope string, r *http.Request) error {
	const op = "lib.auth.ValidateUser"
	log := uv.Log.With(slog.String("op", op))

	var user *user.User
	err := user.ValidateUser(uv.Storage, username, password)
	if err != nil {
		log.Warn("user validation error", sl.Err(err))
		return errors.New("user validation error")
	}

	log.Info("username " + username + " successfully validated")

	return nil
}

// ValidateClient validates clientID and secret returning an error if the client credentials are wrong
// Не используется. Заглушка.
func (uv *UserVerifier) ValidateClient(clientID, clientSecret, scope string, r *http.Request) error {
	/*if clientID == "abcdef" && clientSecret == "12345" {
		return nil
	}*/

	//return errors.New("wrong client")
	return nil
}

// ValidateCode validates token ID
// Не используется. Заглушка.
func (uv *UserVerifier) ValidateCode(clientID, clientSecret, code, redirectURI string, r *http.Request) (string, error) {
	return "", nil
}

// AddClaims provides additional claims to the token
func (uv *UserVerifier) AddClaims(tokenType oauth.TokenType, credential, tokenID, scope string, r *http.Request) (map[string]string, error) {
	const op = "lib.auth.AddClaims"
	log := uv.Log.With(slog.String("op", op))

	claims := make(map[string]string, 1)

	var user *user.User
	user_id, err := user.GetUserID(uv.Storage, credential) // credential contain username
	if err != nil {
		log.Warn("token claims error", sl.Err(err))
		return claims, errors.New("token claims error")
	}

	claims["user_id"] = strconv.Itoa(user_id)
	log.Info("token claims successfully added")

	return claims, nil
}

// AddProperties provides additional information to the token response
// Не используется. Заглушка.
func (uv *UserVerifier) AddProperties(tokenType oauth.TokenType, credential, tokenID, scope string, r *http.Request) (map[string]string, error) {
	props := make(map[string]string, 0)
	//props["customer_name"] = "Gopher"
	return props, nil
}

// ValidateTokenID validates token ID
func (uv *UserVerifier) ValidateTokenID(tokenType oauth.TokenType, credential, tokenID, refreshTokenID string) error {
	const op = "lib.auth.ValidateTokenID"
	log := uv.Log.With(slog.String("op", op))

	var refreshToken *user.RefreshToken
	err := refreshToken.ValidateRefreshTokenID(uv.Storage, credential, refreshTokenID) // credential contain username
	if err != nil {
		log.Warn("token ID validation error", sl.Err(err))
		return errors.New("token ID validation error")
	}

	log.Info("token ID successfully validated")

	return nil
}

// StoreTokenID saves the token id generated for the user
func (uv *UserVerifier) StoreTokenID(tokenType oauth.TokenType, credential, tokenID, refreshTokenID string) error {
	const op = "lib.auth.StoreTokenID"
	log := uv.Log.With(slog.String("op", op))

	var refreshToken *user.RefreshToken
	err := refreshToken.StoreRefreshTokenID(uv.Storage, credential, refreshTokenID) // credential contain username
	if err != nil {
		log.Warn("token ID storing error", sl.Err(err))
		return errors.New("token ID storing error")
	}

	log.Info("token ID successfully stored in DB")

	return nil
}
