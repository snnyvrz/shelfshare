package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/repository"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
	"gorm.io/gorm"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

type CreateBookRequest struct {
	Title       string      `json:"title" binding:"required"`
	AuthorID    uuid.UUID   `json:"author_id" binding:"required,uuid4"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at" swaggertype:"string" example:"2025-11-24"`
}

type UpdateBookRequest struct {
	Title       *string     `json:"title" binding:"omitempty,min=1"`
	AuthorID    *uuid.UUID  `json:"author_id" binding:"omitempty,uuid4"`
	Description *string     `json:"description" binding:"omitempty,max=2000"`
	PublishedAt *model.Date `json:"published_at" swaggertype:"string" example:"2025-11-24"`
}
type BookResponse struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Author      AuthorResponse `json:"author"`
	Description string         `json:"description"`
	PublishedAt *model.Date    `json:"published_at,omitempty" swaggertype:"string" example:"2025-11-24"`
	CreatedAt   model.Date     `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt   model.Date     `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
}

type BookSummaryResponse struct {
	ID          uuid.UUID   `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at,omitempty" swaggertype:"string" example:"2025-11-24"`
	CreatedAt   model.Date  `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt   model.Date  `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
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
		ID:    b.ID,
		Title: b.Title,
		Author: AuthorResponse{
			ID:   b.Author.ID,
			Name: b.Author.Name,
			Bio:  b.Author.Bio,
			CreatedAt: model.Date{
				Time: b.Author.CreatedAt,
			},
			UpdatedAt: model.Date{
				Time: b.Author.UpdatedAt,
			},
		},
		Description: b.Description,
		PublishedAt: pub,
		CreatedAt:   model.Date{Time: b.CreatedAt},
		UpdatedAt:   model.Date{Time: b.UpdatedAt},
	}
}

func toBookSummaryResponse(b model.Book) BookSummaryResponse {
	var pub *model.Date
	if b.PublishedAt != nil && !b.PublishedAt.IsZero() {
		pub = &model.Date{Time: *b.PublishedAt}
	}

	return BookSummaryResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		PublishedAt: pub,
		CreatedAt:   model.Date{Time: b.CreatedAt},
		UpdatedAt:   model.Date{Time: b.UpdatedAt},
	}
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
		AuthorID:    req.AuthorID,
		Description: req.Description,
		PublishedAt: pubAt,
	}

	ctx := c.Request.Context()

	if err := h.repo.Create(ctx, &book); err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_CREATE_FAILED",
			"failed to create book",
		)
		return
	}

	// Ensure Author is loaded (repo handles Preload)
	created, err := h.repo.FindByID(ctx, book.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_FETCH_FAILED",
			"failed to fetch created book",
		)
		return
	}

	c.JSON(http.StatusCreated, toBookResponse(*created))
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
	ctx := c.Request.Context()

	books, err := h.repo.List(ctx)
	if err != nil {
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

	ctx := c.Request.Context()

	book, err := h.repo.FindByID(ctx, bookID)
	if err != nil {
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

	c.JSON(http.StatusOK, toBookResponse(*book))
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

	ctx := c.Request.Context()

	book, err := h.repo.FindByID(ctx, bookID)
	if err != nil {
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

	if req.Title == nil && req.AuthorID == nil &&
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
	if req.AuthorID != nil {
		book.AuthorID = *req.AuthorID
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

	if err := h.repo.Update(ctx, book); err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_UPDATE_FAILED",
			"failed to update book",
		)
		return
	}

	updated, err := h.repo.FindByID(ctx, book.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_FETCH_FAILED",
			"failed to fetch updated book",
		)
		return
	}

	c.JSON(http.StatusOK, toBookResponse(*updated))
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

	ctx := c.Request.Context()

	err = h.repo.Delete(ctx, bookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(c, http.StatusNotFound,
				"BOOK_NOT_FOUND",
				"book not found",
			)
			return
		}

		writeError(c, http.StatusInternalServerError,
			"BOOK_DELETE_FAILED",
			"failed to delete book",
		)
		return
	}

	c.Status(http.StatusNoContent)
}
