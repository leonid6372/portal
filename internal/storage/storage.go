package storage

import "errors"

var (
	ErrWrongAuth = errors.New("login or password is incorrect")
	ErrEmptyCart = errors.New("cart is empty")
)
