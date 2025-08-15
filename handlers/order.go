package handlers

import (
	"net/http"
	"strconv"

	"ecommerce-app/models"
	"ecommerce-app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db             *gorm.DB
	paymentService *services.PaymentService
}

func NewOrderHandler(db *gorm.DB, paymentService *services.PaymentService) *OrderHandler {
	return &OrderHandler{
		db:             db,
		paymentService: paymentService,
	}
}

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's cart
	var cart models.Cart
	if err := h.db.Preload("CartItems.Product").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
		return
	}

	if len(cart.CartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	// Calculate total and validate stock
	var totalAmount float64

	for _, item := range cart.CartItems {
		if item.Product.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for product: " + item.Product.Name})
			return
		}
		totalAmount += item.Product.Price * float64(item.Quantity)
	}

	// Create order
	order := models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
	}

	if err := h.db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order items and update stock
	for _, item := range cart.CartItems {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		}
		if err := h.db.Create(&orderItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
			return
		}

		// Update product stock
		item.Product.Stock -= item.Quantity
		if err := h.db.Save(&item.Product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	// Create Stripe checkout session
	checkoutReq := services.CreateCheckoutSessionRequest{
		Amount:      int64(totalAmount * 100), // Convert to cents
		Currency:    "usd",
		SuccessURL:  "http://localhost:3000/order/success?session_id={CHECKOUT_SESSION_ID}",
		CancelURL:   "http://localhost:3000/order/cancel",
		OrderID:     strconv.FormatUint(uint64(order.ID), 10),
		Description: "Order payment",
	}

	checkoutResp, err := h.paymentService.CreateCheckoutSession(checkoutReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment session"})
		return
	}

	// Update order with session ID
	order.StripeSessionID = checkoutResp.SessionID
	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	// Clear cart
	if err := h.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order": order,
		"checkout_url": checkoutResp.URL,
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var orders []models.Order
	if err := h.db.Preload("OrderItems.Product").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")
	id, err := strconv.ParseUint(orderID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	if err := h.db.Preload("OrderItems.Product").Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *OrderHandler) GetMyProductOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var orders []models.Order
	if err := h.db.Preload("OrderItems.Product").Preload("User").
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Joins("JOIN products ON order_items.product_id = products.id").
		Where("products.user_id = ?", userID).
		Group("orders.id").
		Order("orders.created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")
	id, err := strconv.ParseUint(orderID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns any product in this order
	var order models.Order
	if err := h.db.Preload("OrderItems.Product").
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Joins("JOIN products ON order_items.product_id = products.id").
		Where("orders.id = ? AND products.user_id = ?", id, userID).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or you don't have permission to update it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Update order status
	order.Status = req.Status
	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *OrderHandler) ConfirmPayment(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// Find order by session ID
	var order models.Order
	if err := h.db.Where("stripe_session_id = ?", sessionID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Update order status to paid
	order.Status = models.OrderStatusPaid
	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed successfully", "order": order})
}
