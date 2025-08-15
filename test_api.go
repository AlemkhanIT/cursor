package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-app/handlers"
	"ecommerce-app/models"
	"ecommerce-app/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	
	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
		&models.Review{},
		&models.Message{},
	)
	if err != nil {
		panic("failed to migrate database")
	}
	
	return db
}

func TestRegisterEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	emailService := services.NewEmailService()
	authHandler := handlers.NewAuthHandler(db, emailService)

	router := gin.New()
	router.POST("/register", authHandler.Register)

	// Test data
	registerData := map[string]interface{}{
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
	}

	jsonData, _ := json.Marshal(registerData)

	// Create request
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "user_id")
}

func TestLoginEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	emailService := services.NewEmailService()
	authHandler := handlers.NewAuthHandler(db, emailService)

	router := gin.New()
	router.POST("/login", authHandler.Login)

	// Test data
	loginData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(loginData)

	// Create request
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions - should fail because user doesn't exist
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProductCreation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	productHandler := handlers.NewProductHandler(db)

	router := gin.New()
	router.POST("/products", productHandler.CreateProduct)

	// Test data
	productData := map[string]interface{}{
		"name":        "Test Product",
		"description": "A test product",
		"price":       99.99,
		"stock":       10,
		"category":    "Electronics",
	}

	jsonData, _ := json.Marshal(productData)

	// Create request
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions - should fail because no authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test functions are defined above
