package storageHandler

import "errors"

var (
	ErrItemUnavailable  = errors.New("Item with selected item_id is not available for order")
	ErrPlaceUnavailable = errors.New("Place with selected place_id is not available for reservation")
)
