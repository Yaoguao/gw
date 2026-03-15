package storage

import "errors"

var (
	ErrWalletNotFound = errors.New("wallet not found")

	ErrInsufficientFunds = errors.New("insufficient funds")

	ErrUserNotFound = errors.New("user not found")
)
