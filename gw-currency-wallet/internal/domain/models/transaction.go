package models

import (
	"time"

	"github.com/google/uuid"
)

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      OperationType
	Currency  string
	Amount    int64
	CreatedAt time.Time
}
