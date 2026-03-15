package models

import (
	"time"

	"github.com/google/uuid"
)

type Exchange struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	FromCurrency string
	ToCurrency   string
	AmountFrom   int64
	AmountTo     int64
	Rate         int64
	CreatedAt    time.Time
}
