package mssql

import (
	"database/sql"
	"fmt"

	"portal/internal/config"

	_ "github.com/denisenkom/go-mssqldb"
)

type Storage struct {
	DB *sql.DB
}

func New(cfg config.SQL) (*Storage, error) {
	const op = "storage.mssql.New" // Имя текущей функции для логов и ошибок

	DB, err := sql.Open(cfg.MssqlDriver, cfg.MssqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: DB}, nil
}
