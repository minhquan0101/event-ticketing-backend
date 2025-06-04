package controllers

import (
	"context"
	"net/http"

	"event-ticketing/config"
	"event-ticketing/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DTO để trả về đơn hàng chi tiết
type OrderDetail struct {
	OrderID      primitive.ObjectID `json:"order_id"`
	EventName    string             `json:"event_name"`
	Quantity     int                `json:"quantity"`
	TicketPrice  float64            `json:"ticket_price"`
	TotalPrice   float64            `json:"total_price"`
	Status       string             `json:"status"`
	EventDate    string             `json:"event_date"`
	PurchaseTime string             `json:"purchase_time"`
}

// GetMyOrders godoc
// @Summary Xem đơn hàng của tôi
// @Description Lấy danh sách các đơn hàng đã đặt của người dùng hiện tại
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Success 200 {array} controllers.OrderDetail
// @Failure 500 {object} map[string]string
// @Router /api/orders/my [get]
func GetMyOrders(c *gin.Context) {
	db := config.GetDB()
	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	// 1. Lấy tất cả đơn hàng của user
	orderCursor, err := db.Collection("orders").Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy đơn hàng"})
		return
	}
	defer orderCursor.Close(context.TODO())

	var orders []models.Order
	if err := orderCursor.All(context.TODO(), &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý dữ liệu"})
		return
	}

	// 2. Tìm ticket + event tương ứng cho từng đơn
	var result []OrderDetail

	for _, order := range orders {
		// Tìm ticket
		var ticket models.Ticket
		err := db.Collection("tickets").FindOne(context.TODO(), bson.M{"_id": order.TicketID}).Decode(&ticket)
		if err != nil {
			continue // bỏ qua nếu không tìm thấy
		}

		// Tìm event
		var event models.Event
		err = db.Collection("events").FindOne(context.TODO(), bson.M{"_id": ticket.EventID}).Decode(&event)
		if err != nil {
			continue
		}

		result = append(result, OrderDetail{
			OrderID:      order.ID,
			EventName:    event.Name,
			Quantity:     ticket.Quantity,
			TicketPrice:  event.TicketPrice,
			TotalPrice:   order.TotalPrice,
			Status:       order.Status,
			EventDate:    event.Date.Format("2006-01-02 15:04"),
			PurchaseTime: ticket.PurchaseTime.Format("2006-01-02 15:04"),
		})
	}

	c.JSON(http.StatusOK, result)
}
