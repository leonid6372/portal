package postgres

import (
	"database/sql"
	_ "encoding/json"
	"fmt"

	_ "github.com/go-playground/validator"
)

const (
	qrGetStoreList = `SELECT jsonb_agg(item) FROM item`
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.NewStorage" // Имя текущей функции для логов и ошибок

	db, err := sql.Open("postgres", storagePath) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetStoreListJSON() (*string, error) {
	const op = "storage.postgres.GetStoreLisJSON" // Имя текущей функции для логов и ошибок

	qrResult, err := s.db.Query(qrGetStoreList)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var storeListJSON []byte
	if err := qrResult.Scan(&storeListJSON); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	res := string(storeListJSON)

	return &res, nil
}
