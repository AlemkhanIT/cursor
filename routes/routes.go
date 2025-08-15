package routes

import (
	"ecommerce-app/handlers"
	"ecommerce-app/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(
	router *gin.Engine,
	db *gorm.DB,
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
	cartHandler *handlers.CartHandler,
	reviewHandler *handlers.ReviewHandler,
	messageHandler *handlers.MessageHandler,
	websocketService *services.WebSocketService,
	authMiddleware gin.HandlerFunc,
) {
	// Add database to context
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/confirm-email", authHandler.ConfirmEmail)
			auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(authMiddleware)
		{
			// Product routes
			products := protected.Group("/products")
			{
				products.GET("", productHandler.GetProducts)
				products.GET("/:id", productHandler.GetProduct)
				products.POST("", productHandler.CreateProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
				products.GET("/my", productHandler.GetMyProducts)
			}

			// Cart routes
			cart := protected.Group("/cart")
			{
				cart.GET("", cartHandler.GetCart)
				cart.POST("/add", cartHandler.AddToCart)
				cart.PUT("/items/:id", cartHandler.UpdateCartItem)
				cart.DELETE("/items/:id", cartHandler.RemoveFromCart)
				cart.DELETE("", cartHandler.ClearCart)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetOrders)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.GET("/my-products", orderHandler.GetMyProductOrders)
				orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
			}

			// Review routes
			reviews := protected.Group("/reviews")
			{
				reviews.POST("", reviewHandler.CreateReview)
				reviews.GET("/product/:id", reviewHandler.GetProductReviews)
				reviews.PUT("/:id", reviewHandler.UpdateReview)
				reviews.DELETE("/:id", reviewHandler.DeleteReview)
				reviews.GET("/my", reviewHandler.GetMyReviews)
			}

			// Message routes
			messages := protected.Group("/messages")
			{
				messages.POST("", messageHandler.SendMessage)
				messages.GET("/conversations", messageHandler.GetConversations)
				messages.GET("/conversation/:user_id", messageHandler.GetConversation)
				messages.GET("/unread-count", messageHandler.GetUnreadCount)
				messages.PUT("/:id/read", messageHandler.MarkAsRead)
				messages.DELETE("/:id", messageHandler.DeleteMessage)
			}
		}

		// Payment confirmation (no authentication required for webhook)
		api.GET("/orders/confirm-payment", orderHandler.ConfirmPayment)
	}

	// WebSocket route
	router.GET("/ws", func(c *gin.Context) {
		// Extract user info from query params or headers
		userID := c.Query("user_id")
		username := c.Query("username")
		
		if userID == "" || username == "" {
			c.JSON(400, gin.H{"error": "user_id and username are required"})
			return
		}
		
		// Convert userID to uint
		uid, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid user_id"})
			return
		}
		
		websocketService.HandleWebSocket(c.Writer, c.Request, uint(uid), username)
	})
}
