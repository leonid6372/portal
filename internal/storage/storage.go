package storage

import "errors"

var (
	ErrItemUnavailable = errors.New("Item with selected item_id is not available for order")
	ErrPassword        = errors.New("Password do not match")
)
