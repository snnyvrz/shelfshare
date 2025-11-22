package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
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
	e.GET("/books/:id", h.GetBookByID)
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
	if !validation.BindAndValidateJSON(c, &req) {
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

	responses := make([]BookResponse, 0, len(books))
	for _, b := range books {
		responses = append(responses, toBookResponse(b))
	}

	c.JSON(http.StatusOK, responses)
}

func (h *BookHandler) GetBookByID(c *gin.Context) {
	idParam := c.Param("id")

	bookID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	var book model.Book

	if err := h.db.First(&book, "id = ?", bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "book not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch book",
		})
		return
	}

	c.JSON(http.StatusOK, toBookResponse(book))
}
