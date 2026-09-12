package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/repository"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/validation"
	"gorm.io/gorm"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" && pgErr.ConstraintName == "fk_authors_books" {
				writeError(c, http.StatusBadRequest,
					"AUTHOR_NOT_FOUND",
					"author does not exist",
				)
				return
			}
		}

		writeError(c, http.StatusInternalServerError,
			"BOOK_CREATE_FAILED",
			"failed to create book",
		)
		return
	}

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
// @Param        page            query     int     false  "Page number"      default(1) minimum(1)
// @Param        page_size       query     int     false  "Items per page"   default(20) minimum(1) maximum(100)
// @Param        sort            query     string  false  "Sort field and direction" Enums(created_at_desc,created_at_asc,title_asc,title_desc,published_at_desc,published_at_asc)
// @Param        q               query     string  false  "Full-text search on title and description"
// @Param        author_id       query     string  false  "Filter by author ID (UUID)"
// @Param        published_after query     string  false  "Filter: published_at >= YYYY-MM-DD" example(2015-01-01)
// @Param        published_before query    string  false  "Filter: published_at <= YYYY-MM-DD" example(2020-12-31)
// @Success      200  {object}   ListBooksResponse
// @Failure      400  {object}  validation.ErrorResponse   "Invalid query parameters"
// @Failure      500  {object}  validation.ErrorResponse   "Internal server error"
// @Router       /books [get]
func (h *BookHandler) ListBooks(c *gin.Context) {
	ctx := c.Request.Context()

	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	sort := c.DefaultQuery("sort", "created_at_desc")

	query := c.Query("q")

	var authorIDPtr *uuid.UUID
	if authorStr := c.Query("author_id"); authorStr != "" {
		id, err := uuid.Parse(authorStr)
		if err != nil {
			writeError(c, http.StatusBadRequest,
				"INVALID_AUTHOR_ID",
				"author_id must be a valid UUID",
			)
			return
		}
		authorIDPtr = &id
	}

	pubAfter, err := parseDateQuery(c, "published_after")
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"INVALID_PUBLISHED_AFTER",
			"published_after must be in format YYYY-MM-DD",
		)
		return
	}

	pubBefore, err := parseDateQuery(c, "published_before")
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"INVALID_PUBLISHED_BEFORE",
			"published_before must be in format YYYY-MM-DD",
		)
		return
	}

	params := repository.BookListParams{
		Page:      page,
		PageSize:  pageSize,
		Sort:      sort,
		Query:     query,
		AuthorID:  authorIDPtr,
		PubAfter:  pubAfter,
		PubBefore: pubBefore,
	}

	result, err := h.repo.List(ctx, params)
	if err != nil {
		writeError(c, http.StatusInternalServerError,
			"BOOK_LIST_FAILED",
			"failed to fetch books",
		)
		return
	}

	responses := make([]Book, 0, len(result.Books))
	for _, b := range result.Books {
		responses = append(responses, toBookResponse(b).Data)
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = int((result.Total + int64(params.PageSize) - 1) / int64(params.PageSize))
	}

	c.JSON(http.StatusOK, toListBooksResponse(responses, params.Page, params.PageSize, result.Total, totalPages))
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

	if err := h.repo.Delete(c.Request.Context(), bookID); err != nil {
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
