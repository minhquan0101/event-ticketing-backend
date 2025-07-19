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

	socketio "github.com/googollee/go-socket.io"

	_ "event-ticketing/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load biến môi trường
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Không thể load file .env, sẽ dùng biến hệ thống")
	}

	// Kết nối DB & Redis
	config.ConnectDB()
	config.ConnectRedis()
	if config.GetDB() == nil {
		log.Fatal("❌ Không thể khởi tạo MongoDB – kiểm tra ConnectDB()")
	}

	// Khởi tạo router Gin
	r := gin.Default()

	// ✅ Cấu hình CORS – mở cho localhost và domain frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
    "http://localhost:5173",                    // local dev
    "https://client.minhquan.site",            // ✅ domain thật
    "https://event-ticketing-frontend.onrender.com",
	},

		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Tài liệu Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Đăng ký các API routes
	routes.RegisterRoutes(r)

	// ✅ Khởi tạo server socket.io – phiên bản mới chỉ trả về 1 giá trị
	server := socketio.NewServer(nil)
	config.SocketServer = server

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("🟢 Socket client connected:", s.ID())
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("🔴 Socket client disconnected:", s.ID(), "Lý do:", reason)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("⚠️ Socket error:", e)
	})

	// Gắn socket server vào Gin
	r.GET("/socket.io/*any", gin.WrapH(server))
	r.POST("/socket.io/*any", gin.WrapH(server))

	// Phục vụ static cho ảnh
	r.Static("/static", "./static")

	// Lấy PORT từ env hoặc mặc định 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Chạy server
	log.Println("🚀 Server chạy tại http://localhost:" + port)
	log.Println("📚 Swagger tại     http://localhost:" + port + "/swagger/index.html")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Không thể khởi chạy server:", err)
	}
}
