package postgres

import (
	"database/sql"
	_ "encoding/json"
	"fmt"

	"portal/internal/config"

	_ "github.com/go-playground/validator"
	_ "github.com/lib/pq"
)

const (
	qrGetStoreList = `SELECT jsonb_agg(item) FROM item`
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.SQLStorage) (*Storage, error) {
	const op = "storage.postgres.NewStorage" // Имя текущей функции для логов и ошибок

	db, err := sql.Open(cfg.SQLDriver, cfg.SQLInfo) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetStoreList() (*string, error) {
	const op = "storage.postgres.GetStoreList" // Имя текущей функции для логов и ошибок

	qrResult, err := s.db.Query(qrGetStoreList)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var storeList string
	for qrResult.Next() {
		if err := qrResult.Scan(&storeList); err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
		}
	}

	return &storeList, nil
}
