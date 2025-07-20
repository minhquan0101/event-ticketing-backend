// @title Event Ticketing API
// @version 1.0
// @description API backend quản lý sự kiện và đặt vé bằng Golang
// @contact.name Dự án nhóm - sử dụng với ChatGPT
// @contact.email your_email@example.com
// @host api.minhquan.site
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

	// ✅ Cấu hình CORS cho các domain frontend
	r.Use(cors.New(cors.Config{
    AllowOriginFunc: func(origin string) bool {
        // Chấp nhận các domain cụ thể
        return origin == "http://localhost:5173" ||
            origin == "https://client.minhquan.site" ||
            origin == "https://event-ticketing-frontend.onrender.com"
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
}))


	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Đăng ký API routes
	routes.RegisterRoutes(r)

	// Cấu hình socket.io
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

	r.GET("/socket.io/*any", gin.WrapH(server))
	r.POST("/socket.io/*any", gin.WrapH(server))

	// Serve static files
	r.Static("/static", "./static")

	// Cổng
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ✅ Quan trọng: dùng 0.0.0.0 thay vì localhost
	log.Println("🚀 Server chạy tại http://0.0.0.0:" + port)
	log.Println("📚 Swagger tại     http://0.0.0.0:" + port + "/swagger/index.html")
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("❌ Không thể khởi chạy server:", err)
	}
}
