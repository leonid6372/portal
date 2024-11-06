package user

import (
	"fmt"
	storageHandler "portal/internal/storage"
	"portal/internal/storage/mssql"
	"portal/internal/storage/postgres"
)

const (
	qrNewUser                   = `INSERT INTO "user" (role, balance, username, full_name, position, department) VALUES ($5, 0, $1, $2, $3, $4) RETURNING user_id;`
	qrGetUserFullName           = `SELECT _Fld7254 FROM [10295].[dbo].[_InfoRg7251] WHERE _Fld7252 = $1;`
	qrGetUserInfo               = `SELECT full_name, position, department FROM "user" WHERE username = $1;`
	qrGetRole                   = `SELECT "role" FROM "user" WHERE username = $1;`
	qrGetPassByUsername         = `SELECT "password" FROM "user" WHERE username = $1;`
	qrGetUserIDByUsername       = `SELECT user_id FROM "user" WHERE username = $1;`
	qrGetUserById               = `SELECT "1c" FROM "user" WHERE user_id = $1;`
	qrGetUsernameByUserID       = `SELECT username FROM "user" WHERE user_id = $1;`
	qrGetRefreshTokenIDByUserID = `SELECT refresh_token_id FROM refresh_token WHERE user_id = $1;`
	qrStoreRefreshTokenID       = `INSERT INTO refresh_token (user_id, refresh_token_id)
								   VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE
								   SET refresh_token_id = EXCLUDED.refresh_token_id;`
)

type User struct {
	UserID     int    `json:"user_id,omitempty"`
	Data1C     string `json:"data_1c,omitempty"`
	Username   string `json:"username,omitempty"`
	FullName   string `json:"full_name,omitempty"`
	Position   string `json:"position,omitempty"`
	Department string `json:"department,omitempty"`
	Balance    int    `json:"balance,omitempty"`
	Password   string `json:"password,omitempty"`
	Role       int    `json:"role,omitempty"`
}

func (u *User) NewUser(storage *postgres.Storage, username, fullName, position, department string, role int) error {
	const op = "storage.postgres.entities.user.NewUser"

	err := storage.DB.QueryRow(qrNewUser, username, fullName, position, department, role).Scan(&u.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

/*func (u *User) GetUserById(storage *postgres.Storage) error {
	const op = "storage.postgres.entities.user.GetUserById" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetUserById, u.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong user_id", op)
	}

	if err := qrResult.Scan(&u.Data1C); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}*/

func (u *User) ValidateUser(storage *postgres.Storage, storage1C *mssql.Storage, username, password string) error {
	const op = "storage.postgres.entities.user.ValidateUser"

	// Проверяем username в БД 1С
	stmt, err := storage1C.DB.Prepare(qrGetUserFullName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	qrResult, err := stmt.Query(username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	// TO DO: Включить проверку на пароль

	/*qrResult, err := storage.DB.Query(qrGetPassByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}*/

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}

	/*if err := qrResult.Scan(&u.Password); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if password != u.Password {
		return fmt.Errorf("%s: wrong password", op)
	}*/

	return nil
}

func (u *User) GetUserID(storage *postgres.Storage, username string) error {
	const op = "storage.postgres.entities.user.GetUserID"

	qrResult, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: %w", op, storageHandler.ErrUserIDDoesNotExist)
	}

	if err := qrResult.Scan(&u.UserID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) GetUsername(storage *postgres.Storage, userID int) error {
	const op = "storage.postgres.entities.user.GetUsername"

	err := storage.DB.QueryRow(qrGetUsernameByUserID, userID).Scan(&u.Username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) GetUserInfo(storage *postgres.Storage, username string) error {
	const op = "storage.postgres.entities.user.GetUserInfo"

	stmt, err := storage.DB.Prepare(qrGetUserInfo)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&u.FullName, &u.Position, &u.Department)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) GetFullNameByUsername(storage1C *mssql.Storage, username string) error {
	const op = "storage.postgres.entities.user.GetFullNameByUsername"

	stmt, err := storage1C.DB.Prepare(qrGetUserFullName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&u.FullName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) GetRoleByUsername(storage *postgres.Storage, username string) error {
	const op = "storage.postgres.entities.user.GetRoleByUsername"

	err := storage.DB.QueryRow(qrGetRole, username).Scan(&u.Role)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type RefreshToken struct {
	UserID         int    `json:"user_id,omitempty"`
	RefreshTokenID string `json:"refresh_token_id,omitempty"`
}

func (r *RefreshToken) ValidateRefreshTokenID(storage *postgres.Storage, username, refreshTokenID string) error {
	const op = "storage.postgres.entities.user.ValidateRefreshTokenID"

	qrResult1, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult1.Close()

	if !qrResult1.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	var userID int
	if err := qrResult1.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	qrResult2, err := storage.DB.Query(qrGetRefreshTokenIDByUserID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult2.Close()

	if !qrResult2.Next() {
		return fmt.Errorf("%s: wrong userID", op)
	}
	var correctRefreshTokenID string
	if err := qrResult2.Scan(&correctRefreshTokenID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if refreshTokenID != correctRefreshTokenID {
		return fmt.Errorf("%s: wrong refresh token ID", op)
	}

	return nil
}

func (r *RefreshToken) StoreRefreshTokenID(storage *postgres.Storage, username, refreshTokenID string) error {
	const op = "storage.postgres.entities.user.StoreRefreshTokenID"

	qrResult1, err := storage.DB.Query(qrGetUserIDByUsername, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult1.Close()

	if !qrResult1.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}
	var userID int
	if err := qrResult1.Scan(&userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = storage.DB.Exec(qrStoreRefreshTokenID, userID, refreshTokenID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
