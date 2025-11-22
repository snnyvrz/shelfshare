package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type BookResponse struct {
	ID          uuid.UUID   `json:"id"`
	Title       string      `json:"title"`
	Author      string      `json:"author"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at,omitempty"`
	CreatedAt   model.Date  `json:"created_at"`
	UpdatedAt   model.Date  `json:"updated_at"`
}

func (h *BookHandler) RegisterRoutes(e *gin.Engine) {
	e.POST("/books", h.CreateBook)
	e.GET("/books", h.ListBooks)
}

func toBookResponse(b model.Book) BookResponse {
	var pub *model.Date
	if b.PublishedAt != nil && !b.PublishedAt.IsZero() {
		pub = &model.Date{Time: *b.PublishedAt}
	}

	return BookResponse{
		ID:          b.ID,
		Title:       b.Title,
		Author:      b.Author,
		Description: b.Description,
		PublishedAt: pub,
		CreatedAt:   model.Date{Time: b.CreatedAt},
		UpdatedAt:   model.Date{Time: b.UpdatedAt},
	}
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

	var pubAt *time.Time
	if req.PublishedAt != nil && !req.PublishedAt.Time.IsZero() {
		t := req.PublishedAt.Time
		pubAt = &t
	}

	book := model.Book{
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		PublishedAt: pubAt,
	}

	if err := h.db.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create book",
		})
		return
	}

	c.JSON(http.StatusCreated, toBookResponse(book))
}

func (h *BookHandler) ListBooks(c *gin.Context) {
	var books []model.Book

	if err := h.db.Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch books",
		})
		return
	}

	responses := make([]BookResponse, len(books))
	for i, b := range books {
		responses[i] = toBookResponse(b)
	}

	c.JSON(http.StatusOK, responses)
}
