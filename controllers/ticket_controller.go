package controllers

import (
	"context"
	"net/http"
	"time"

	"event-ticketing/config"
	"event-ticketing/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaceOrder godoc
// @Summary Đặt vé và tạo đơn hàng
// @Description Người dùng đặt vé cho sự kiện, hệ thống sẽ tạo ticket và order
// @Tags Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body object{event_id=string,quantity=int} true "Thông tin đặt vé"
// @Success 200 {object} map[string]interface{message=string,total_price=float64}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tickets/order [post]
func PlaceOrder(c *gin.Context) {  // Đặt vé sự kiện + tính tiền tự động
	var req struct {
		EventID  string `json:"event_id"`
		Quantity int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	eventID, err := primitive.ObjectIDFromHex(req.EventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID sự kiện không hợp lệ"})
		return
	}

	// Lấy sự kiện từ DB
	eventCollection := config.GetDB().Collection("events")
	var event models.Event
	err = eventCollection.FindOne(context.TODO(), bson.M{"_id": eventID}).Decode(&event)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự kiện"})
		return
	}

	if event.AvailableTickets < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không đủ vé còn lại"})
		return
	}

	// Tính tổng tiền
	totalPrice := float64(req.Quantity) * event.TicketPrice

	// Trừ vé
	_, err = eventCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": eventID},
		bson.M{"$inc": bson.M{"available_tickets": -req.Quantity}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật số vé"})
		return
	}

	// Lấy user ID từ JWT
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// Tạo ticket
	ticket := models.Ticket{
		UserID:       userID,
		EventID:      eventID,
		Quantity:     req.Quantity,
		PurchaseTime: time.Now(),
	}
	ticketResult, err := config.GetDB().Collection("tickets").InsertOne(context.TODO(), ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo vé"})
		return
	}

	// Tạo đơn hàng với tổng tiền
	order := models.Order{
		UserID:     userID,
		TicketID:   ticketResult.InsertedID.(primitive.ObjectID),
		Status:     "pending",
		CreatedAt:  time.Now(),
		TotalPrice: totalPrice,
	}
	_, err = config.GetDB().Collection("orders").InsertOne(context.TODO(), order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo đơn hàng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Đặt vé và tạo đơn hàng thành công",
		"total_price":  totalPrice,
		"quantity":     req.Quantity,
		"event_name":   event.Name,
		"ticket_price": event.TicketPrice,
	})
}
