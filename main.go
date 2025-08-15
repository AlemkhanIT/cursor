package main

import (
	"log"
	"os"

	"ecommerce-app/config"
	"ecommerce-app/handlers"
	"ecommerce-app/middleware"
	"ecommerce-app/models"
	"ecommerce-app/routes"
	"ecommerce-app/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate database
	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Review{},
		&models.Message{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize services
	emailService := services.NewEmailService()
	paymentService := services.NewPaymentService()
	websocketService := services.NewWebSocketService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, emailService)
	productHandler := handlers.NewProductHandler(db)
	orderHandler := handlers.NewOrderHandler(db, paymentService)
	cartHandler := handlers.NewCartHandler(db)
	reviewHandler := handlers.NewReviewHandler(db)
	messageHandler := handlers.NewMessageHandler(db, websocketService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware()

	// Setup router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Setup routes
	routes.SetupRoutes(router, db, authHandler, productHandler, orderHandler, cartHandler, reviewHandler, messageHandler, websocketService, authMiddleware)

	// Start WebSocket hub
	go websocketService.StartHub()

	// Get port from environment
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
