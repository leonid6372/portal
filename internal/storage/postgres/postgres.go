package postgres

import (
	"database/sql"
	"fmt"

	"portal/internal/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func New(cfg config.SQL) (*Storage, error) {
	const op = "storage.postgres.New" // Имя текущей функции для логов и ошибок

	DB, err := sql.Open(cfg.PostgresDriver, cfg.PostgresInfo) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: DB}, nil
}
