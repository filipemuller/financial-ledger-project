package models

import "errors"

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountExists        = errors.New("account already exists")
	ErrInvalidAccountID     = errors.New("invalid account ID")
	ErrNegativeBalance      = errors.New("balance cannot be negative")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidAmount        = errors.New("amount must be positive")
	ErrSameAccount          = errors.New("cannot transfer to same account")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrDuplicateIdempotency = errors.New("duplicate idempotency key")
)
