package storageHandler

import "errors"

var (
	ErrCartDoesNotExist = errors.New("cart does not exist")
	ErrPageInOutOfRange = errors.New("page in out of range")
)
