package postgres

import (
	"database/sql"
	"fmt"

	"portal/internal/config"

	_ "github.com/lib/pq"
)

const (
	qrGetShopList = `SELECT jsonb_agg(item) FROM item`
	qrAddCartItem = `SELECT add_cart_item($1, $2)`
	qrGetUser     = `SELECT "password" FROM "user" WHERE login = $1`
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

func (s *Storage) GetShopList() (string, error) {
	const op = "storage.postgres.GetShopList" // Имя текущей функции для логов и ошибок

	qrResult, err := s.db.Query(qrGetShopList)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var shopList string
	for qrResult.Next() {
		if err := qrResult.Scan(&shopList); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return shopList, nil
}

func (s *Storage) AddCartItem(item_id, quantity int) error {
	const op = "storage.postgres.AddCartItem"

	_, err := s.db.Query(qrAddCartItem, item_id, quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUser(login, password string) (bool, error) {
	return true, nil
}
