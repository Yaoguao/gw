package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LargeTranslation struct {
	ID bson.ObjectID `bson:"_id,omitempty"`

	TranslationID string    `bson:"translation_id" json:"translation_id"`
	FromUserID    string    `bson:"from_user_id" json:"from_user_id"`
	ToUserID      string    `bson:"to_user_id" json:"to_user_id"`
	Amount        int64     `bson:"amount" json:"amount"`
	Currency      string    `bson:"currency" json:"currency"`
	CreatedAt     time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	ReceivedAt    time.Time `bson:"received_at,omitempty" json:"received_at,omitempty"`
}
