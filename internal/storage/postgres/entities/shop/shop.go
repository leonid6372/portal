package shop

import (
	"fmt"
	"portal/internal/storage/postgres"
	"time"

	storageHandler "portal/internal/storage"
)

const (

	qrNewCart                   = `INSERT INTO cart(user_id, is_active) VALUES ($1, true);`
	qrGetItems                  = `SELECT * FROM item;`
	qrGetInCartItems            = `SELECT in_cart_item_id, item_id, quantity FROM in_active_cart_item WHERE cart_id = $1;`
	qrGetActiveCartID           = `SELECT cart_id FROM cart WHERE user_id = $1 AND is_active = true;`
	qrGetIsAvailable            = `SELECT is_available FROM item WHERE item_id = $1;`
	qrDeleteItem                = `DELETE FROM item WHERE item_id = $1;`
	qrDeleteInCartItemsByCartID = `DELETE FROM in_cart_item WHERE cart_id = $1;`
	qrDeleteInCartItem          = `DELETE FROM in_cart_item WHERE in_cart_item_id = $1;`
	qrUpdateInCartItem          = `UPDATE in_cart_item SET quantity = $1 WHERE in_cart_item_id = $2;`
	qrUpdateCartToInactive      = `UPDATE cart SET is_active = false and "date" = localtimestamp WHERE user_id = $1 and is_active = true;`
	qrNewInCartItem             = `INSERT INTO in_cart_item(item_id, quantity, cart_id)
					  			   VALUES($1, $2, $3) ON CONFLICT (item_id, cart_id) DO
								   UPDATE SET quantity = in_cart_item.quantity + $2;`
)

type Item struct {
	ItemID      int    `json:"item_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price,omitempty"`
	PhotoPath   string `json:"photo_path,omitempty"`
	IsAvailable bool   `json:"is_available,omitempty"`
}

func (i *Item) DeleteItem(storage *postgres.Storage, itemID int) error {
	const op = "storage.postgres.entities.shop.DeleteItem"

	_, err := storage.DB.Exec(qrDeleteItem, itemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (i *Item) GetIsAvailable(storage *postgres.Storage, itemID int) error {
	const op = "storage.postgres.entities.shop.GetIsAvailable"

	qrResult, err := storage.DB.Query(qrGetIsAvailable, itemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !qrResult.Next() {
		return fmt.Errorf("%s: %w", op, "item is not exist")
	}

	if err := qrResult.Scan(&i.IsAvailable); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (i *Item) GetItems(storage *postgres.Storage) ([]Item, error) {
	const op = "storage.postgres.entities.shop.GetItems"

	qrResult, err := storage.DB.Query(qrGetItems)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var is []Item

	for qrResult.Next() {
		if err := qrResult.Scan(&i.ItemID, &i.Name, &i.Description, &i.Price, &i.PhotoPath, &i.IsAvailable); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		is = append(is, *i)
	}

	return is, nil
}

type InCartItem struct {
	InCartItemID int `json:"in_cart_item_id,omitempty"`
	CartID       int `json:"cart_id,omitempty"`
	ItemID       int `json:"item_id,omitempty"`
	Quantity     int `json:"quantity,omitempty"`
}

func (ici *InCartItem) NewInCartItem(storage *postgres.Storage, itemID, quantity, cartID int) error {
	const op = "storage.postgres.entities.shop.NewInCartItem"

	_, err := storage.DB.Exec(qrNewInCartItem, itemID, quantity, cartID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ici *InCartItem) DeleteInCartItem(storage *postgres.Storage, inCartItemID int) error {
	const op = "storage.postgres.entities.shop.DeleteInCartItem"

	_, err := storage.DB.Exec(qrDeleteInCartItem, inCartItemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ici *InCartItem) UpdateInCartItem(storage *postgres.Storage, inCartItemID, quantity int) error {
	const op = "storage.postgres.entities.shop.UpdateInCartItem"

	_, err := storage.DB.Exec(qrUpdateInCartItem, quantity, inCartItemID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ici *InCartItem) GetInCartItems(storage *postgres.Storage, cartID int) ([]InCartItem, error) {
	const op = "storage.postgres.entities.shop.GetInCartItems"

	qrResult, err := storage.DB.Query(qrGetInCartItems, cartID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var icis []InCartItem

	for qrResult.Next() {
		var ici InCartItem
		if err := qrResult.Scan(&ici.InCartItemID, &ici.ItemID, &ici.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		icis = append(icis, ici)
	}

	return icis, nil
}

type Cart struct {
	CartID   int       `json:"cart_id,omitempty"`
	UserID   int       `json:"user_id,omitempty"`
	IsActive bool      `json:"is_active,omitempty"`
	Date     time.Time `json:"date,omitempty"`
}

func (c *Cart) UpdateCartToInactive(storage *postgres.Storage, userID int) error {
	const op = "storage.postgres.entities.shop.UpdateCartToInactive"

	_, err := storage.DB.Exec(qrUpdateCartToInactive, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cart) GetActiveCartID(storage *postgres.Storage, userID int) error {
	const op = "storage.postgres.entities.shop.GetActiveCartID"

	qrResult, err := storage.DB.Query(qrGetActiveCartID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !qrResult.Next() {
		return fmt.Errorf("%s: %w", op, storageHandler.ErrCartDoesNotExist)
	}

	if err := qrResult.Scan(&c.CartID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cart) NewCart(storage *postgres.Storage, userID int) error {
	const op = "storage.postgres.entities.shop.NewCart"

	_, err := storage.DB.Exec(qrNewCart, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cart) EmptyCart(storage *postgres.Storage, userID int) error {
	const op = "storage.postgres.entities.shop.EmptyCart"

	_, err := storage.DB.Exec(qrDeleteInCartItemsByCartID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
