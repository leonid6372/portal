package Shop

import (
	"fmt"
	db "portal/internal/storage/postgres"
)

type Item struct {
	ItemID      int    `json:"item_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	PhotoPath   string `json:"photo_path"`
	IsActive    bool   `json:"is_active" validate:"required"`
}

type InCartItem struct {
	InCartItemID int
	ItemId       int
	Quantity     int
}

const (
	qrGetShopList = `SELECT jsonb_agg(item) FROM item`
	qrAddCartItem = `INSERT INTO in_cart_item (item_id, quantity) VALUES ($1, $2)`
)

func (i *Item) GetShopList(db *db.Storage) (string, error) {
	const op = "storage.postgres.entities.GetShopList" // Имя текущей функции для логов и ошибок

	qrResult, err := db.DB.Query(qrGetShopList)
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

func (c *InCartItem) Add(db *db.Storage, Item_id, Quantity int) error {
	const op = "storage.postgres.Add"

	_, err := db.DB.Query(qrAddCartItem, Item_id, Quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
