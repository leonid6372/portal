package oauth

import (
	"errors"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	storageHandler "portal/internal/storage"
	ldapServer "portal/internal/storage/ldap"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"
	"portal/internal/structs/roles"
)

// UserVerifier provides user credentials verifier for testing. Все методы этой структуры нужны для удовлетворения условиям NewBearerServer
type UserVerifier struct {
	Storage *postgres.Storage
	//Storage1C  *mssql.Storage
	LDAPServer *ldapServer.LDAPServer
	Log        *slog.Logger
}

// ValidateUser validates username and password returning an error if the user credentials are wrong
func (uv *UserVerifier) ValidateUser(username, password string, r *http.Request) (int, error) {
	const op = "lib.oauth.ValidateUser"
	log := uv.Log.With(slog.String("op", op))

	// Проверка пользователя через LDAP. Получаем DN пользваотеля для авторизации через него (проверка пароля), также получаем инфо о пользователе
	userInfo, err := uv.LDAPServer.GetUserInfo(username)
	if err != nil {
		log.Error(op, "failed to get user info", sl.Err(err))
		return 0, errors.New("failed to get user info: " + err.Error())
	}
	err = uv.LDAPServer.LDAPConn.Bind(userInfo[0], password)
	if err != nil {
		log.Error(op, "LDAP user validation error", sl.Err(err))
		return 0, errors.New("LDAP user validation error: " + err.Error())
	}

	// Проверка пользователя по базе 1С
	/*var u user.User
	err = u.ValidateUser(uv.Storage, uv.Storage1C, username, password)
	if err != nil {
		log.Warn("user validation error", sl.Err(err))
		return 0, errors.New("user validation error: " + err.Error())
	}*/

	// Get user id
	var u user.User
	u.Role = roles.User
	err = u.GetUserID(uv.Storage, username)
	if err != nil {
		// Если ошибка не об отсутствии user_id, то выход по стнадартной ошибке БД
		if !errors.As(err, &storageHandler.ErrUserIDDoesNotExist) {
			log.Error(op, "failed to get user id", sl.Err(err))
			return 0, errors.New("token claims error: " + err.Error())
		}
		isMemberOfGroup, err := uv.LDAPServer.IsUserMemberOf(username, "KD Heads employees and specialists")
		if err != nil {
			log.Error(op, "failed to check is user member of group", sl.Err(err))
			return 0, errors.New("token claims error: " + err.Error())
		}
		if !isMemberOfGroup {
			u.Role = roles.UserWithOutReservation
		}
		// Если ошибка выше была об отсутствии user_id, то создаем user_id для пользователя и получаем его в u.UserID
		if err := u.NewUser(uv.Storage, username, userInfo[1], userInfo[2], userInfo[3], u.Role); err != nil {
			log.Error(op, "failed to create user in postgres", sl.Err(err))
			return 0, errors.New("token claims error: " + err.Error())
		}
	}

	err = u.GetRoleByUsername(uv.Storage, username)
	if err != nil {
		log.Warn("failed to get user role", sl.Err(err))
		return 0, errors.New("failed to get user role: " + err.Error())
	}

	log.Info("username " + username + " successfully validated")

	return u.Role, nil
}

// AddClaims provides additional claims to the token
func (uv *UserVerifier) AddClaims(credential, tokenID string, scope int, r *http.Request) (map[string]int, error) {
	const op = "lib.oauth.AddClaims"
	log := uv.Log.With(slog.String("op", op))

	claims := make(map[string]int, 1) // 1 - количество параметров claims в токене

	// Get user id
	var u user.User
	u.Role = roles.User
	err := u.GetUserID(uv.Storage, credential)
	if err != nil {
		// Если ошибка не об отсутствии user_id, то выход по стнадартной ошибке БД
		if !errors.As(err, &storageHandler.ErrUserIDDoesNotExist) {
			log.Error(op, "failed to get user id", sl.Err(err))
			return claims, errors.New("token claims error: " + err.Error())
		}
		// Проверка пользователя через LDAP. Получаем DN пользваотеля для авторизации через него (проверка пароля), также получаем инфо о пользователе
		userInfo, err := uv.LDAPServer.GetUserInfo(credential)
		if err != nil {
			log.Error(op, "failed to get user info", sl.Err(err))
			return claims, errors.New("failed to get user info: " + err.Error())
		}
		isMemberOfGroup, err := uv.LDAPServer.IsUserMemberOf(credential, "KD Heads employees and specialists")
		if err != nil {
			log.Error(op, "failed to check is user member of group", sl.Err(err))
			return claims, errors.New("token claims error: " + err.Error())
		}
		if !isMemberOfGroup {
			u.Role = roles.UserWithOutReservation
		}
		// Если ошибка выше была об отсутствии user_id, то создаем user_id для пользователя и получаем его в u.UserID
		if err := u.NewUser(uv.Storage, credential, userInfo[1], userInfo[2], userInfo[3], u.Role); err != nil {
			log.Error(op, "failed to create user in postgres", sl.Err(err))
			return claims, errors.New("token claims error: " + err.Error())
		}
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
		log.Error(op, "token ID storing error", sl.Err(err))
		return errors.New("token ID storing error: " + err.Error())
	}

	log.Info("token ID successfully stored in DB")

	return nil
}
