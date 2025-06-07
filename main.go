// @title Event Ticketing API
// @version 1.0
// @description API backend quản lý sự kiện và đặt vé bằng Golang
// @contact.name Dự án nhóm - sử dụng với ChatGPT
// @contact.email your_email@example.com
// @host localhost:8080
// @BasePath /

package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"event-ticketing/config"
	"event-ticketing/routes"

	_ "event-ticketing/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load biến môi trường
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Không thể load file .env, sẽ dùng biến hệ thống")
	}

	// Kết nối MongoDB và Redis
	config.ConnectDB()
	config.ConnectRedis()
	if config.GetDB() == nil {
		log.Fatal("❌ Không thể khởi tạo MongoDB – kiểm tra ConnectDB()")
	}

	// Tạo router Gin
	r := gin.Default()

	// ✅ Cấu hình CORS đầy đủ để cho phép frontend gửi Authorization
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Route tài liệu Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Đăng ký API route chính
	routes.RegisterRoutes(r)

	// Lấy PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Chạy server
	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("❌ Không thể khởi chạy server:", err)
	}

	log.Println("🚀 Server chạy tại http://localhost:" + port)
	log.Println("📚 Swagger tại     http://localhost:" + port + "/swagger/index.html")
}
