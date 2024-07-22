package storageHandler

import "errors"

var (
	ErrCartDoesNotExist   = errors.New("cart does not exist")
	ErrUserIDDoesNotExist = errors.New("user id doesn not exist")
	ErrPageInOutOfRange   = errors.New("page in out of range")
)
