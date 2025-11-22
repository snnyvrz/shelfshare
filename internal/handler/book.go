package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/go-book-crud-gin/internal/model"
	"gorm.io/gorm"
)

type BookHandler struct {
	db *gorm.DB
}

func NewBookHandler(db *gorm.DB) *BookHandler {
	return &BookHandler{db: db}
}

type CreateBookRequest struct {
	Title       string      `json:"title" binding:"required"`
	Author      string      `json:"author" binding:"required"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at"`
}

func (h *BookHandler) RegisterRoutes(e *gin.Engine) {
	e.POST("/books", h.CreateBook)
}

func (h *BookHandler) CreateBook(c *gin.Context) {
	var req CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	book := model.Book{
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		PublishedAt: req.PublishedAt,
	}

	if err := h.db.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create book",
		})
		return
	}

	c.JSON(http.StatusCreated, book)
}
