package shop

import (
	"database/sql"
	"fmt"
	"portal/internal/storage/postgres"
	"time"
)

const (
	qrGetShopList    = `SELECT jsonb_agg(item) FROM item`
	qrAddCartItem    = `INSERT INTO in_cart_item (item_id, quantity, cart_id) VALUES ($1, $2, $3)`
	qrGetCartData    = `SELECT jsonb_agg(json_object('in_cart_item_id': in_cart_item_id, 'item_id': item_id, 'quantity': quantity)) FROM UserAvailableCart where is_active = true and user_id = ($1)`
	qrGetActualCart  = `SELECT cart_id FROM cart WHERE user_id = ($1) AND is_active = true`
	qrCreateCart     = `INSERT INTO cart(user_id, is_active) VALUES ($1, true)`
	qrDropCartItem   = `DELETE FROM in_cart_item WHERE in_cart_item_id = ($1)`
	qrUpdateCartItem = `UPDATE in_cart_item SET quantity = ($1) WHERE in_cart_item_id = ($2)`
	qrOrder          = `UPDATE cart SET is_active = false and "date" = localtimestamp WHERE user_id = ($1) and is_active = true`
)

type Item struct {
	ItemID      int    `json:"itemId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	PhotoPath   string `json:"photoPath"`
	IsAvailable bool   `json:"isAvailable"`
}

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

func (ici *InCartItem) AddCartItem(storage *postgres.Storage, itemID, quantity int) error {
	const op = "storage.postgres.entities.shop.AddCartItem"

	var c *Cart
	CartID, err := c.CreateCart(storage, 1) // сделать чтобы юзер id вытаскивался

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = storage.DB.Exec(qrAddCartItem, itemID, quantity, CartID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ici *InCartItem) DropCartItem(storage *postgres.Storage, inCartItemID int) error {
	const op = "storage.postgres.entities.shop.DropCartItem"

	_, err := storage.DB.Exec(qrDropCartItem, inCartItemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ici *InCartItem) UpdateCartItem(storage *postgres.Storage, inCartItemID, quantity int) error {
	const op = "storage.postgres.entities.shop.UpdateCartItem"

	_, err := storage.DB.Exec(qrUpdateCartItem, quantity, inCartItemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

type Cart struct {
	CartID   int       `json:"cart_id"`
	UserID   int       `json:"user_id"`
	IsActive bool      `json:"is_active"`
	Date     time.Time `json:"date"`
}

func (c *Cart) CreateCart(storage *postgres.Storage, userID int) (int, error) {
	const op = "storage.postgres.entities.shop.CreateCart"

	qrResult, err := c.GetActualCart(storage, userID)
	if !qrResult.Next() {
		_, err = storage.DB.Exec(qrCreateCart, userID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		qrResult, err = c.GetActualCart(storage, userID)
	}

	var CartID int
	if err = qrResult.Scan(&CartID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return CartID, nil
}

func (c *Cart) GetActualCart(storage *postgres.Storage, userID int) (*sql.Rows, error) {
	const op = "storage.postgres.entities.shop.GetActualCart"
	qrResult, err := storage.DB.Query(qrGetActualCart, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return qrResult, nil
}

func (c *Cart) GetCartData(storage *postgres.Storage, userID int) ([]byte, error) {
	const op = "storage.postgres.entities.shop.GetCartData"

	_, err := c.CreateCart(storage, userID)
	qrResult, err := storage.DB.Query(qrGetCartData, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var cartItemList []byte
	for qrResult.Next() {
		if err := qrResult.Scan(&cartItemList); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return cartItemList, nil
}

func (c *Cart) Order(storage *postgres.Storage, userID int) error {
	{
		const op = "storage.postgres.entities.shop.UpdateCartItem"

		_, err := storage.DB.Exec(qrOrder, userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
}

func (c *Cart) GetUserOrderList(storage *postgres.Storage, userID int) (string, error) {
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
