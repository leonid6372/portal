package oauth

import (
	"errors"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"
)

// UserVerifier provides user credentials verifier for testing. Все методы этой структуры нужны для удовлетворения условиям NewBearerServer
type UserVerifier struct {
	Storage *postgres.Storage
	Log     *slog.Logger
}

// ValidateUser validates username and password returning an error if the user credentials are wrong
func (uv *UserVerifier) ValidateUser(username, password, scope string, r *http.Request) error {
	const op = "lib.oauth.ValidateUser"
	log := uv.Log.With(slog.String("op", op))

	var user user.User
	err := user.ValidateUser(uv.Storage, username, password)
	if err != nil {
		log.Warn("user validation error", sl.Err(err))
		return errors.New("user validation error: " + err.Error())
	}

	log.Info("username " + username + " successfully validated")

	return nil
}

// AddClaims provides additional claims to the token
func (uv *UserVerifier) AddClaims(credential, tokenID, scope string, r *http.Request) (map[string]int, error) {
	const op = "lib.oauth.AddClaims"
	log := uv.Log.With(slog.String("op", op))

	claims := make(map[string]int, 1) // 1 - количество параметров claims в токене

	var u user.User
	err := u.GetUserID(uv.Storage, credential) // credential contains username
	if err != nil {
		log.Error("token claims error", sl.Err(err))
		return claims, errors.New("token claims error: " + err.Error())
	}

	claims["user_id"] = u.UserID
	log.Info("token claims successfully added")

	return claims, nil
}

// ValidateTokenID validates token ID
func (uv *UserVerifier) ValidateTokenID(credential, tokenID, refreshTokenID string) error {
	const op = "lib.oauth.ValidateTokenID"
	log := uv.Log.With(slog.String("op", op))

	var refreshToken *user.RefreshToken
	err := refreshToken.ValidateRefreshTokenID(uv.Storage, credential, refreshTokenID) // credential contains username
	if err != nil {
		log.Warn("token ID validation error", sl.Err(err))
		return errors.New("token ID validation error: " + err.Error())
	}

	log.Info("token ID successfully validated")

	return nil
}

// StoreTokenID saves the token id generated for the user
func (uv *UserVerifier) StoreTokenID(credential, tokenID, refreshTokenID string) error {
	const op = "lib.oauth.StoreTokenID"
	log := uv.Log.With(slog.String("op", op))

	var refreshToken *user.RefreshToken
	err := refreshToken.StoreRefreshTokenID(uv.Storage, credential, refreshTokenID) // credential contains username
	if err != nil {
		log.Error("token ID storing error", sl.Err(err))
		return errors.New("token ID storing error: " + err.Error())
	}

	log.Info("token ID successfully stored in DB")

	return nil
}
