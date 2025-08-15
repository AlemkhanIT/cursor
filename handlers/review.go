package handlers

import (
	"net/http"
	"strconv"

	"ecommerce-app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewHandler struct {
	db *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{db: db}
}

type CreateReviewRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product exists
	var product models.Product
	if err := h.db.Where("id = ? AND is_active = ?", req.ProductID, true).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	// Check if user already reviewed this product
	var existingReview models.Review
	if err := h.db.Where("user_id = ? AND product_id = ?", userID, req.ProductID).First(&existingReview).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this product"})
		return
	}

	// Create review
	review := models.Review{
		UserID:    userID,
		ProductID: req.ProductID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := h.db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review": review})
}

func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productID := c.Param("id")
	id, err := strconv.ParseUint(productID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var reviews []models.Review
	if err := h.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name")
	}).Where("product_id = ?", id).Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	reviewID := c.Param("id")
	id, err := strconv.ParseUint(reviewID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var review models.Review
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&review).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		return
	}

	// Update review
	review.Rating = req.Rating
	review.Comment = req.Comment

	if err := h.db.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	reviewID := c.Param("id")
	id, err := strconv.ParseUint(reviewID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Review{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var reviews []models.Review
	if err := h.db.Preload("Product").Where("user_id = ?", userID).Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
