package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"event-ticketing/config"
	"event-ticketing/models"
	"event-ticketing/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
)

// Register godoc
// @Summary Đăng ký tài khoản
// @Description Đăng ký tài khoản mới và gửi mã xác nhận qua email
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body models.User true "Thông tin người dùng"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func Register(c *gin.Context) {  // dang ki
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	userCollection := config.GetDB().Collection("users")

	// Kiểm tra email đã tồn tại chưa
	count, _ := userCollection.CountDocuments(context.TODO(), bson.M{"email": input.Email})
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email đã tồn tại"})
		return
	}

	// Hash mật khẩu
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi mã hóa mật khẩu"})
		return
	}

	// Tạo mã xác nhận 6 số
	verifyCode := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Cập nhật dữ liệu user
	input.Password = hashedPassword
	input.Role = "user"
	input.IsVerified = false
	input.VerifyCode = verifyCode

	// Gửi mã xác nhận qua Gmail
	err = utils.SendVerifyCode(input.Email, verifyCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể gửi mã xác nhận"})
		return
	}

	// Lưu vào MongoDB
	_, err = userCollection.InsertOne(context.TODO(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo tài khoản"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Đăng ký thành công. Vui lòng kiểm tra email để xác nhận."})
}

// Login godoc
// @Summary Đăng nhập
// @Description Đăng nhập và trả về JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body models.User true "Thông tin đăng nhập"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/login [post]
func Login(c *gin.Context) {  // Đăng nhập
	var input models.User
	var foundUser models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	userCollection := config.GetDB().Collection("users")

	err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai email hoặc mật khẩu"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, foundUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai email hoặc mật khẩu"})
		return
	}

	if !foundUser.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Email chưa được xác minh"})
		return
	}

	// Tạo JWT token
	claims := jwt.MapClaims{
		"user_id": foundUser.ID.Hex(),
		"email":   foundUser.Email,
		"role":    foundUser.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không tạo được token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// VerifyEmail godoc
// @Summary Xác minh email
// @Description Gửi mã xác nhận để xác thực tài khoản email
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body object{email=string,code=string} true "Email và mã xác nhận"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/verify-email [post]
func VerifyEmail(c *gin.Context) {// Xác minh email
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	userCollection := config.GetDB().Collection("users")

	filter := bson.M{"email": req.Email}
	var user models.User
	err := userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	if user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Tài khoản đã xác minh"})
		return
	}

	if user.VerifyCode != req.Code {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã xác nhận không đúng"})
		return
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"is_verified": true, "verify_code": ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xác minh tài khoản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xác minh thành công!"})
}
