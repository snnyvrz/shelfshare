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

// CreateBookRequest represents the payload for creating a book.
type CreateBookRequest struct {
	Title       string      `json:"title" binding:"required"`
	Author      string      `json:"author" binding:"required"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at"`
}

// UpdateBookRequest represents the payload for partially updating a book.
type UpdateBookRequest struct {
	Title       *string     `json:"title" binding:"omitempty,min=1"`
	Author      *string     `json:"author" binding:"omitempty,min=1"`
	Description *string     `json:"description" binding:"omitempty,max=2000"`
	PublishedAt *model.Date `json:"published_at"`
}

// BookResponse is the response representation of a book.
type BookResponse struct {
	ID          uuid.UUID   `json:"id"`
	Title       string      `json:"title"`
	Author      string      `json:"author"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at,omitempty"`
	CreatedAt   model.Date  `json:"created_at"`
	UpdatedAt   model.Date  `json:"updated_at"`
}

func (h *BookHandler) RegisterRoutes(r *gin.RouterGroup) {
	books := r.Group("/books")
	{
		books.GET("", h.ListBooks)
		books.GET("/:id", h.GetBookByID)
		books.PATCH("/:id", h.UpdateBook)
		books.DELETE("/:id", h.DeleteBook)
		books.POST("", h.CreateBook)
	}
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

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, validation.ErrorResponse{
		Code:    code,
		Message: message,
		Errors:  nil,
	})
}

// CreateBook godoc
// @Summary      Create a book
// @Description  Create a new book with title, author, description and optional published date
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        payload  body      CreateBookRequest          true  "Book to create"
// @Success      201      {object}  BookResponse
// @Failure      400      {object}  validation.ErrorResponse   "Validation error"
// @Failure      500      {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books [post]
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
		writeError(c, http.StatusInternalServerError,
			"BOOK_CREATE_FAILED",
			"failed to create book",
		)
		return
	}

	c.JSON(http.StatusCreated, toBookResponse(book))
}

// ListBooks godoc
// @Summary      List books
// @Description  Get all books
// @Tags         books
// @Produce      json
// @Success      200  {array}   BookResponse
// @Failure      500  {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books [get]
func (h *BookHandler) ListBooks(c *gin.Context) {
	var books []model.Book

	if err := h.db.Find(&books).Error; err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_LIST_FAILED",
			"failed to fetch books",
		)
		return
	}

	responses := make([]BookResponse, 0, len(books))
	for _, b := range books {
		responses = append(responses, toBookResponse(b))
	}

	c.JSON(http.StatusOK, responses)
}

// GetBookByID godoc
// @Summary      Get a book by ID
// @Description  Get a single book by its UUID
// @Tags         books
// @Produce      json
// @Param        id   path      string  true  "Book ID (UUID)"
// @Success      200  {object}  BookResponse
// @Failure      400  {object}  validation.ErrorResponse   "Invalid ID"
// @Failure      404  {object}  validation.ErrorResponse   "Book not found"
// @Failure      500  {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books/{id} [get]
func (h *BookHandler) GetBookByID(c *gin.Context) {
	idParam := c.Param("id")

	bookID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"INVALID_BOOK_ID",
			"invalid book id",
		)
		return
	}

	var book model.Book

	if err := h.db.First(&book, "id = ?", bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(c, http.StatusNotFound,
				"BOOK_NOT_FOUND",
				"book not found",
			)
			return
		}

		writeError(c, http.StatusInternalServerError,
			"BOOK_FETCH_FAILED",
			"failed to fetch book",
		)
		return
	}

	c.JSON(http.StatusOK, toBookResponse(book))
}

// UpdateBook godoc
// @Summary      Update a book
// @Description  Partially update a book by its UUID
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "Book ID (UUID)"
// @Param        payload  body      UpdateBookRequest   true  "Fields to update"
// @Success      200      {object}  BookResponse
// @Failure      400      {object}  validation.ErrorResponse   "Invalid ID or payload"
// @Failure      404      {object}  validation.ErrorResponse   "Book not found"
// @Failure      500      {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books/{id} [patch]
func (h *BookHandler) UpdateBook(c *gin.Context) {
	idParam := c.Param("id")

	bookID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"INVALID_BOOK_ID",
			"invalid book id",
		)
		return
	}

	var book model.Book
	if err := h.db.First(&book, "id = ?", bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(c, http.StatusNotFound,
				"BOOK_NOT_FOUND",
				"book not found",
			)
			return
		}

		writeError(c, http.StatusInternalServerError,
			"BOOK_FETCH_FAILED",
			"failed to fetch book",
		)
		return
	}

	var req UpdateBookRequest
	if !validation.BindAndValidateJSON(c, &req) {
		return
	}

	if req.Title == nil && req.Author == nil &&
		req.Description == nil && req.PublishedAt == nil {
		writeError(c, http.StatusBadRequest,
			"NO_FIELDS_TO_UPDATE",
			"at least one field must be provided to update",
		)
		return
	}

	if req.Title != nil {
		book.Title = *req.Title
	}
	if req.Author != nil {
		book.Author = *req.Author
	}
	if req.Description != nil {
		book.Description = *req.Description
	}
	if req.PublishedAt != nil {
		if req.PublishedAt.Time.IsZero() {
			book.PublishedAt = nil
		} else {
			t := req.PublishedAt.Time
			book.PublishedAt = &t
		}
	}

	if err := h.db.Save(&book).Error; err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_UPDATE_FAILED",
			"failed to update book",
		)
		return
	}

	c.JSON(http.StatusOK, toBookResponse(book))
}

// DeleteBook godoc
// @Summary      Delete a book
// @Description  Delete a book by its UUID
// @Tags         books
// @Produce      json
// @Param        id   path      string  true  "Book ID (UUID)"
// @Success      204  {string}  string  "No content"
// @Failure      400  {object}  validation.ErrorResponse   "Invalid ID"
// @Failure      404  {object}  validation.ErrorResponse   "Book not found"
// @Failure      500  {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books/{id} [delete]
func (h *BookHandler) DeleteBook(c *gin.Context) {
	idParam := c.Param("id")

	bookID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"INVALID_BOOK_ID",
			"invalid book id",
		)
		return
	}

	result := h.db.Delete(&model.Book{}, "id = ?", bookID)
	if result.Error != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_DELETE_FAILED",
			"failed to delete book",
		)
		return
	}

	if result.RowsAffected == 0 {
		writeError(c, http.StatusNotFound,
			"BOOK_NOT_FOUND",
			"book not found",
		)
		return
	}

	c.Status(http.StatusNoContent)
}
