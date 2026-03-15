package models

import "github.com/google/uuid"

type Wallet struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Currency string
	Balance  int64
}
