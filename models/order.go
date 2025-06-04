package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	TicketID   primitive.ObjectID `bson:"ticket_id" json:"ticket_id"`
	Status     string             `bson:"status" json:"status"` // "pending", "paid", "canceled"
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	TotalPrice float64            `bson:"total_price" json:"total_price"` // optional
}
