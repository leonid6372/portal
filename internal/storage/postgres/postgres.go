package postgres

import (
	"database/sql"
	"fmt"

	"portal/internal/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db *sql.DB
}

func New(cfg config.SQLStorage) (*Storage, error) {
	const op = "storage.postgres.NewStorage" // Имя текущей функции для логов и ошибок

	db, err := sql.Open(cfg.SQLDriver, cfg.SQLInfo) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{Db: db}, nil
}
