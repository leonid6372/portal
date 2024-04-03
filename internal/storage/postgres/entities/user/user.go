package user

import (
	"fmt"
	"portal/internal/storage/postgres"
)

const (
	qrGetPassByUsername         = `SELECT "password" FROM "user" WHERE username = $1`
	qrGetUserIDByUsername       = `SELECT user_id FROM "user" WHERE username = $1`
	qrGetUserById               = `SELECT username, balance, "1c" FROM "user" WHERE user_id = $1`
	qrGetRefreshTokenIDByUserID = `SELECT refresh_token_id FROM refresh_token WHERE user_id = $1`
	qrStoreRefreshTokenID       = `INSERT INTO refresh_token (user_id, refresh_token_id)
								   VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE
								   SET refresh_token_id = EXCLUDED.refresh_token_id`
)

type User struct {
	UserID    int    `json:"userId"`
	Data1C    string `json:"data1c"`
	Username  string `json:"username"`
	Balance   int    `json:"balance"`
	Password  string `json:"password"`
	AccessLVL int    `json:"accessLvl"`
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

	if err := qrResult.Scan(&u.Username, &u.Balance, &u.Data1C); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TO DO: Переписать под ORM
func (u *User) ValidateUser(storage *postgres.Storage, username, password string) error {
	const op = "storage.postgres.entities.user.ValidateUser"

	var correctPassword string
	qrResult, err := storage.DB.Query(qrGetPassByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}

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

	var userID int
	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return 0, fmt.Errorf("%s: wrong username", op)
	}

	if err := qrResult.Scan(&userID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

type RefreshToken struct {
	UserID         int    `json:"userId"`
	RefreshTokenID string `json:"refreshTokenId"`
}

func (r *RefreshToken) ValidateRefreshTokenID(storage *postgres.Storage, username, refreshTokenID string) error {
	const op = "storage.postgres.entities.user.ValidateRefreshTokenID"

	var userID int
	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	if err := qrResult.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var correctRefreshTokenID string
	qrResult, err = storage.DB.Query(qrGetRefreshTokenIDByUserID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong userID", op)
	}
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

	var userID int
	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	if err := qrResult.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = storage.DB.Exec(qrStoreRefreshTokenID, userID, refreshTokenID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
