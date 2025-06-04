package routes

import (
	"event-ticketing/controllers"
	"event-ticketing/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Group API
	api := r.Group("/api")

	// Auth routes (public)
	api.POST("/register", controllers.Register)
	api.POST("/login", controllers.Login)

	// Event routes (public)
	api.GET("/events", controllers.GetAllEvents)
	api.GET("/events/:id", controllers.GetEventByID)
	api.POST("/verify-email", controllers.VerifyEmail)

	// Protected routes (yêu cầu JWT)
	protected := api.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Ticket
		protected.POST("/tickets/order", controllers.PlaceOrder)
		protected.GET("/orders/my", controllers.GetMyOrders)
		// protected.POST("/orders", controllers.CreateOrder)


		// Optional: routes dành cho admin
		protected.POST("/events", controllers.CreateEvent)
		protected.PUT("/events/:id", controllers.UpdateEvent)
		protected.DELETE("/events/:id", controllers.DeleteEvent)
	}
}
