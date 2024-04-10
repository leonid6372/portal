package user

import (
	"fmt"
	"portal/internal/storage/postgres"
)

const (
	qrGetPassByUsername         = `SELECT "password" FROM "user" WHERE username = $1;`
	qrGetUserIDByUsername       = `SELECT user_id FROM "user" WHERE username = $1;`
	qrGetUserById               = `SELECT "1c" FROM "user" WHERE user_id = $1;`
	qrGetRefreshTokenIDByUserID = `SELECT refresh_token_id FROM refresh_token WHERE user_id = $1;`
	qrStoreRefreshTokenID       = `INSERT INTO refresh_token (user_id, refresh_token_id)
								   VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE
								   SET refresh_token_id = EXCLUDED.refresh_token_id;`
)

type User struct {
	UserID    int    `json:"user_id,omitempty"`
	Data1C    string `json:"data_1c,omitempty"`
	Username  string `json:"username,omitempty"`
	Balance   int    `json:"balance,omitempty"`
	Password  string `json:"password,omitempty"`
	AccessLVL int    `json:"access_lvl,omitempty"`
}

func (u *User) GetUserById(storage *postgres.Storage) error {
	const op = "storage.postgres.entities.user.GetUserById" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetUserById, u.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong user_id", op)
	}

	if err := qrResult.Scan(&u.Data1C); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TO DO: Переписать под ORM
func (u *User) ValidateUser(storage *postgres.Storage, username, password string) error {
	const op = "storage.postgres.entities.user.ValidateUser"

	qrResult, err := storage.DB.Query(qrGetPassByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}

	var correctPassword string
	if err := qrResult.Scan(&correctPassword); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if password != correctPassword {
		return fmt.Errorf("%s: wrong password", op)
	}

	return nil
}

// TO DO: Переписать под ORM
func (u *User) GetUserID(storage *postgres.Storage, username string) (int, error) {
	const op = "storage.postgres.entities.user.GetUserID"

	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return 0, fmt.Errorf("%s: wrong username", op)
	}

	var userID int
	if err := qrResult.Scan(&userID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

type RefreshToken struct {
	UserID         int    `json:"user_id,omitempty"`
	RefreshTokenID string `json:"refresh_token_id,omitempty"`
}

func (r *RefreshToken) ValidateRefreshTokenID(storage *postgres.Storage, username, refreshTokenID string) error {
	const op = "storage.postgres.entities.user.ValidateRefreshTokenID"

	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	var userID int
	if err := qrResult.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	qrResult, err = storage.DB.Query(qrGetRefreshTokenIDByUserID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong userID", op)
	}
	var correctRefreshTokenID string
	if err := qrResult.Scan(&correctRefreshTokenID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if refreshTokenID != correctRefreshTokenID {
		return fmt.Errorf("%s: wrong refresh token ID", op)
	}

	return nil
}

func (r *RefreshToken) StoreRefreshTokenID(storage *postgres.Storage, username, refreshTokenID string) error {
	const op = "storage.postgres.entities.user.StoreRefreshTokenID"

	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	var userID int
	if err := qrResult.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = storage.DB.Exec(qrStoreRefreshTokenID, userID, refreshTokenID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
