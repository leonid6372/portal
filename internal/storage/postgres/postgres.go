package postgres

import (
	"database/sql"
	"fmt"

	"portal/internal/config"

	_ "github.com/lib/pq"
)

const (
	qrGetShopList = `SELECT jsonb_agg(item) FROM item`

	qrCheckAvailableQuantity = `SELECT quantity
								  FROM item
								 WHERE item_id = $1`

	qrAddCartItem = `INSERT INTO in_cart_item(item_id, quantity)
					      VALUES ($1, $2)
						   
					 ON CONFLICT (item_id) DO 
					  UPDATE SET quantity = quantity + $2`
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
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var shopList string
	for qrResult.Next() {
		if err := qrResult.Scan(&shopList); err != nil {
			fmt.Println(err)
			return "", fmt.Errorf("%s: prepare statement: %w", op, err)
		}
	}

	return shopList, nil
}

func (s *Storage) AddCartItem(item_id, quantity int) error {

	return nil
}
