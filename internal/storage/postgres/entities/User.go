package entities

import (
	"fmt"
	"portal/internal/config"
	errHandler "portal/internal/storage"
	db "portal/internal/storage/postgres"
)

type User struct {
	userId    int
	userLogin string
	balance   int
}

type UserData struct {
	User
}

const (
	qrGetUserById = `SELECT "login", "balance" FROM "user" WHERE user_id = $1`
	qrUserAuth    = `SELECT "userId", "login", "balance" FROM "user" WHERE login = $1`

	password = "123"
)

func (u *User) GetUserById() (bool, error) {
	const op = "storage.postgres.entities.getUserById" // Имя текущей функции для логов и ошибок
	s, _ := db.New(config.SQLStorage{})
	qrResult, err := s.Db.Query(qrGetUserById, u.userId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if err := qrResult.Scan(&u.userLogin, &u.balance); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (u *User) UserAuth(db *db.Storage, Login string, Password string) (bool, error) {
	const op = "storage.postgres.entities.userAuth" // Имя текущей функции для логов и ошибок
	qrResult, err := db.Db.Query(qrUserAuth, Login)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if err := qrResult.Scan(&u.userId, &u.userLogin, &u.balance); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	fmt.Printf("%s", Password)
	if Password != password {
		return false, fmt.Errorf("%s: %w", op, errHandler.ErrPassword)
	}
	return true, nil
}
