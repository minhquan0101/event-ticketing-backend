package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ticket struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	EventID      primitive.ObjectID `bson:"event_id" json:"event_id"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	PurchaseTime time.Time          `bson:"purchase_time" json:"purchase_time"`
}
