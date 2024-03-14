package User

import (
	"fmt"
	errHandler "portal/internal/storage"
	db "portal/internal/storage/postgres"
)

type User struct {
	UserID    int
	UserLogin string
	Balance   int
}

const (
	qrGetUserById = `SELECT "login", "balance" FROM "user" WHERE user_id = $1`
	qrUserAuth    = `SELECT "user_id", "login", "balance" FROM "user" WHERE login = $1`

	globalPassword = "123"
)

func (u *User) GetUserById(db *db.Storage) (bool, error) {
	const op = "storage.postgres.entities.getUserById" // Имя текущей функции для логов и ошибок
	qrResult, err := db.DB.Query(qrGetUserById, u.UserID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	for qrResult.Next() {
		if err := qrResult.Scan(&u.UserLogin, &u.Balance); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	return true, nil
}

func (u *User) UserAuth(db *db.Storage, login string, password string) (bool, error) {
	const op = "storage.postgres.entities.userAuth" // Имя текущей функции для логов и ошибок
	qrResult, err := db.DB.Query(qrUserAuth, login)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	for qrResult.Next() {
		if err := qrResult.Scan(&u.UserID, &u.UserLogin, &u.Balance); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	if password != globalPassword {
		return false, fmt.Errorf("%s: %w", op, errHandler.ErrPassword)
	}
	return true, nil
}
