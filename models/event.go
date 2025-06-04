package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`                               // thong tin
	Description      string             `bson:"description" json:"description"`
	Location         string             `bson:"location" json:"location"`
	Date             time.Time          `bson:"date" json:"date"`								// ngay gio
	TotalTickets     int                `bson:"total_tickets" json:"total_tickets"`				// tong ve ban
	AvailableTickets int                `bson:"available_tickets" json:"available_tickets"`		// so ve con lai
	TicketPrice      float64            `bson:"ticket_price" json:"ticket_price"`
	ImageURL         string             `bson:"image_url" json:"image_url"`						// anh su kien
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`					// thoi gian su kien
}
