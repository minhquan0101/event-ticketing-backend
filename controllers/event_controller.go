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

// GetAllEvents godoc
// @Summary Lấy danh sách sự kiện
// @Description Trả về tất cả sự kiện trong hệ thống
// @Tags Event
// @Produce json
// @Success 200 {array} models.Event
// @Failure 500 {object} map[string]string
// @Router /api/events [get]
func GetAllEvents(c *gin.Context) { // GET /api/events
	eventCollection := config.GetDB().Collection("events")

	cursor, err := eventCollection.Find(context.TODO(), bson.M{})
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
func GetEventByID(c *gin.Context) { // GET /api/events/:id
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
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.Event true "Thông tin sự kiện"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events [post]
func CreateEvent(c *gin.Context) { // POST /api/events (admin)
	eventCollection := config.GetDB().Collection("events")

	var input models.Event
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền tạo sự kiện"})
		return
	}

	input.CreatedAt = time.Now()
	input.AvailableTickets = input.TotalTickets

	_, err := eventCollection.InsertOne(context.TODO(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo sự kiện"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo sự kiện thành công"})
}

// UpdateEvent godoc
// @Summary Cập nhật sự kiện
// @Description Admin cập nhật thông tin một sự kiện
// @Tags Event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID sự kiện"
// @Param input body models.Event true "Thông tin cập nhật"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events/{id} [put]
func UpdateEvent(c *gin.Context) {// PUT /api/events/:id (admin)
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

	var update models.Event
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	updateMap := bson.M{
		"name":              update.Name,
		"description":       update.Description,
		"location":          update.Location,
		"date":              update.Date,
		"total_tickets":     update.TotalTickets,
		"available_tickets": update.AvailableTickets,
		"image_url":         update.ImageURL,
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
func DeleteEvent(c *gin.Context) {  // DELETE /api/events/:id (admin)
	eventCollection := config.GetDB().Collection("events")

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

	_, err = eventCollection.DeleteOne(context.TODO(), bson.M{"_id": eventID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xoá sự kiện"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xoá sự kiện thành công"})
}
