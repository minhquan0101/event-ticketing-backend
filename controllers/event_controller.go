package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"event-ticketing/config"
	"event-ticketing/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAllEvents godoc
// @Summary Lấy danh sách sự kiện
// @Description Trả về tất cả sự kiện hoặc theo từ khóa tìm kiếm
// @Tags Event
// @Produce json
// @Param search query string false "Từ khóa tìm kiếm theo tên/mô tả"
// @Success 200 {array} models.Event
// @Failure 500 {object} map[string]string
// @Router /api/events [get]
func GetAllEvents(c *gin.Context) {
	eventCollection := config.GetDB().Collection("events")
	search := c.Query("search")

	// Thiết lập filter tìm kiếm
	filter := bson.M{}
	if search != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": search, "$options": "i"}},
				{"description": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	cursor, err := eventCollection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách sự kiện"})
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.Event
	if err = cursor.All(context.TODO(), &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đọc dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, events)
}


// GetEventByID godoc
// @Summary Lấy chi tiết sự kiện
// @Description Lấy thông tin sự kiện theo ID
// @Tags Event
// @Produce json
// @Param id path string true "ID sự kiện"
// @Success 200 {object} models.Event
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/events/{id} [get]
func GetEventByID(c *gin.Context) {
	eventCollection := config.GetDB().Collection("events")

	idParam := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var event models.Event
	err = eventCollection.FindOne(context.TODO(), bson.M{"_id": eventID}).Decode(&event)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự kiện"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// CreateEvent godoc
// @Summary Tạo sự kiện mới
// @Description Admin tạo một sự kiện mới
// @Tags Event
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Tên sự kiện"
// @Param description formData string false "Mô tả"
// @Param location formData string true "Địa điểm"
// @Param date formData string true "Ngày (RFC3339)"
// @Param total_tickets formData int true "Tổng số vé"
// @Param available_tickets formData int true "Số vé khả dụng"
// @Param ticket_price formData number true "Giá vé"
// @Param image formData file true "Ảnh sự kiện"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events [post]
func CreateEvent(c *gin.Context) {
	eventCollection := config.GetDB().Collection("events")

	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền tạo sự kiện"})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	location := c.PostForm("location")
	dateStr := c.PostForm("date")
	totalTickets := c.PostForm("total_tickets")
	availableTickets := c.PostForm("available_tickets")
	ticketPrice := c.PostForm("ticket_price")

	if name == "" || location == "" || dateStr == "" || totalTickets == "" || availableTickets == "" || ticketPrice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập đầy đủ các trường bắt buộc"})
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil || date.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày sự kiện không hợp lệ hoặc đã qua"})
		return
	}

	totalTicketsInt, err := strconv.Atoi(totalTickets)
	if err != nil || totalTicketsInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tổng số vé không hợp lệ"})
		return
	}
	availableTicketsInt, err := strconv.Atoi(availableTickets)
	if err != nil || availableTicketsInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số vé khả dụng không hợp lệ"})
		return
	}
	ticketPriceFloat, err := strconv.ParseFloat(ticketPrice, 64)
	if err != nil || ticketPriceFloat <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Giá vé không hợp lệ"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng gửi file ảnh với trường 'image'"})
		return
	}

	filename := primitive.NewObjectID().Hex() + "_" + file.Filename
	savePath := "static/uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu ảnh"})
		return
	}

	imageURL := "/static/uploads/" + filename

	event := models.Event{
		Name:            name,
		Description:     description,
		Location:        location,
		Date:            date,
		TotalTickets:    totalTicketsInt,
		AvailableTickets: availableTicketsInt,
		TicketPrice:     ticketPriceFloat,
		ImageURL:        imageURL,
		CreatedAt:       time.Now(),
	}

	_, err = eventCollection.InsertOne(context.TODO(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo sự kiện"})
		return
	}

	// 🔥 Phát sự kiện realtime
	config.SocketServer.BroadcastToNamespace("/", "event_created", gin.H{
		"message": "Sự kiện mới được tạo",
		"event":   event,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo sự kiện thành công", "image_url": imageURL})
}

// UpdateEvent godoc
// @Summary Cập nhật sự kiện
// @Description Admin cập nhật thông tin một sự kiện
// @Tags Event
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID sự kiện"
// @Param name formData string true "Tên sự kiện"
// @Param description formData string false "Mô tả"
// @Param location formData string true "Địa điểm"
// @Param date formData string true "Ngày (RFC3339)"
// @Param total_tickets formData int true "Tổng số vé"
// @Param available_tickets formData int true "Số vé khả dụng"
// @Param ticket_price formData number true "Giá vé"
// @Param image formData file false "Ảnh mới"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [put]
func UpdateEvent(c *gin.Context) {
	eventCollection := config.GetDB().Collection("events")

	eventIDParam := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền cập nhật sự kiện"})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	location := c.PostForm("location")
	dateStr := c.PostForm("date")
	totalTickets := c.PostForm("total_tickets")
	availableTickets := c.PostForm("available_tickets")
	ticketPrice := c.PostForm("ticket_price")

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày không hợp lệ, phải ở dạng RFC3339"})
		return
	}

	totalTicketsInt, err := strconv.Atoi(totalTickets)
	if err != nil || totalTicketsInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tổng vé không hợp lệ"})
		return
	}
	availableTicketsInt, err := strconv.Atoi(availableTickets)
	if err != nil || availableTicketsInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số vé khả dụng không hợp lệ"})
		return
	}
	ticketPriceFloat, err := strconv.ParseFloat(ticketPrice, 64)
	if err != nil || ticketPriceFloat <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Giá vé không hợp lệ"})
		return
	}

	updateMap := bson.M{
		"name":              name,
		"description":       description,
		"location":          location,
		"date":              date,
		"total_tickets":     totalTicketsInt,
		"available_tickets": availableTicketsInt,
		"ticket_price":      ticketPriceFloat,
	}

	file, err := c.FormFile("image")
	if err == nil {
		filename := primitive.NewObjectID().Hex() + "_" + file.Filename
		savePath := "static/uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu ảnh mới"})
			return
		}
		imageURL := "/static/uploads/" + filename
		updateMap["image_url"] = imageURL
	}

	_, err = eventCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": eventID},
		bson.M{"$set": updateMap},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật sự kiện"})
		return
	}

	// 🔥 Phát sự kiện realtime
	config.SocketServer.BroadcastToNamespace("/", "event_updated", gin.H{
		"message":  "Sự kiện đã được cập nhật",
		"event_id": eventID.Hex(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật sự kiện thành công"})
}

// DeleteEvent godoc
// @Summary Xoá sự kiện
// @Description Admin xoá một sự kiện
// @Tags Event
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID sự kiện"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [delete]
func DeleteEvent(c *gin.Context) {
	eventCollection := config.GetDB().Collection("events")
	ticketCollection := config.GetDB().Collection("tickets")

	eventIDParam := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xoá sự kiện"})
		return
	}

	count, err := ticketCollection.CountDocuments(context.TODO(), bson.M{"event_id": eventID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kiểm tra vé của sự kiện"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể xoá sự kiện: đã có vé được đặt"})
		return
	}

	_, err = eventCollection.DeleteOne(context.TODO(), bson.M{"_id": eventID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xoá sự kiện"})
		return
	}

	// 🔥 Phát sự kiện realtime
	config.SocketServer.BroadcastToNamespace("/", "event_deleted", gin.H{
		"message":  "Sự kiện đã bị xoá",
		"event_id": eventID.Hex(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Xoá sự kiện thành công"})
}
