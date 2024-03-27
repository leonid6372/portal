package shop

import (
	"fmt"
	"portal/internal/storage/postgres"
)

const (
	qrGetShopList = `SELECT jsonb_agg(item) FROM item`
	qrAddCartItem = `INSERT INTO in_cart_item (item_id, quantity) VALUES ($1, $2)`
)

type Item struct {
	ItemID      int    `json:"itemId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	PhotoPath   string `json:"photoPath"`
	IsAvailable bool   `json:"isAvailable"`
}

// Переписать под ORM
func (i *Item) GetShopList(storage *postgres.Storage) (string, error) {
	const op = "storage.postgres.entities.shop.GetShopList" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetShopList)
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

type InCartItem struct {
	InCartItemID int `json:"inCartItemId"`
	ItemID       int `json:"itemId"`
	Quantity     int `json:"quantity"`
}

func (c *InCartItem) AddCartItem(storage *postgres.Storage, itemID, quantity int) error {
	const op = "storage.postgres.entities.shop.AddCartItem"

	_, err := storage.DB.Exec(qrAddCartItem, itemID, quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
