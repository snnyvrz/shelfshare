package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
	"gorm.io/gorm"
)

type AuthorHandler struct {
	db *gorm.DB
}

func NewAuthorHandler(db *gorm.DB) *AuthorHandler {
	return &AuthorHandler{db: db}
}

type CreateAuthorRequest struct {
	Name string `json:"name" binding:"required,min=1"`
	Bio  string `json:"bio" binding:"omitempty,max=2000"`
}

type UpdateAuthorRequest struct {
	Name *string `json:"name" binding:"omitempty,min=1"`
	Bio  *string `json:"bio" binding:"omitempty,max=2000"`
}

type AuthorResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	Bio       string                `json:"bio"`
	Books     []BookSummaryResponse `json:"books,omitempty"`
	CreatedAt model.Date            `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt model.Date            `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
}

func (h *AuthorHandler) RegisterRoutes(r *gin.RouterGroup) {
	authors := r.Group("/authors")
	{
		authors.POST("", h.CreateAuthor)
		authors.GET("", h.ListAuthors)
		authors.GET("/:id", h.GetAuthorByID)
		authors.PATCH("/:id", h.UpdateAuthor)
		authors.DELETE("/:id", h.DeleteAuthor)
	}
}

func toAuthorResponse(a model.Author) AuthorResponse {
	books := make([]BookSummaryResponse, 0, len(a.Books))
	for _, b := range a.Books {
		books = append(books, toBookSummaryResponse(b))
	}

	return AuthorResponse{
		ID:        a.ID,
		Name:      a.Name,
		Bio:       a.Bio,
		Books:     books,
		CreatedAt: model.Date{Time: a.CreatedAt},
		UpdatedAt: model.Date{Time: a.UpdatedAt},
	}
}

// CreateAuthor godoc
// @Summary      Create an author
// @Description  Create a new author with name and optional bio
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        payload  body      CreateAuthorRequest        true  "Author to create"
// @Success      201      {object}  AuthorResponse
// @Failure      400      {object}  validation.ErrorResponse   "Validation error"
// @Failure      500      {object}  validation.ErrorResponse   "Internal server error"
// @Router       /authors [post]
func (h *AuthorHandler) CreateAuthor(c *gin.Context) {
	var req CreateAuthorRequest
	if !validation.BindAndValidateJSON(c, &req) {
		return
	}

	author := model.Author{
		Name: req.Name,
		Bio:  req.Bio,
	}

	if err := h.db.Create(&author).Error; err != nil {
		writeError(c, http.StatusInternalServerError,
			"AUTHOR_CREATE_FAILED",
			"failed to create author",
		)
		return
	}

	c.JSON(http.StatusCreated, toAuthorResponse(author))
}

// ListAuthors godoc
// @Summary      List authors
// @Description  Get a list of all authors
// @Tags         authors
// @Accept       json
// @Produce      json
// @Success      200  {array}   AuthorResponse
// @Failure      500  {object}  validation.ErrorResponse   "Internal server error"
// @Router       /authors [get]
func (h *AuthorHandler) ListAuthors(c *gin.Context) {
	var authors []model.Author

	if err := h.db.Preload("Books").Order("created_at DESC").Find(&authors).Error; err != nil {
		writeError(c, http.StatusInternalServerError,
			"AUTHOR_LIST_FAILED",
			"failed to list authors",
		)
		return
	}

	res := make([]AuthorResponse, 0, len(authors))
	for _, a := range authors {
		res = append(res, toAuthorResponse(a))
	}

	c.JSON(http.StatusOK, res)
}

// GetAuthorByID godoc
// @Summary      Get author by ID
// @Description  Get a single author by its ID
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        id   path      string                    true  "Author ID (UUID)"
// @Success      200  {object}  AuthorResponse
// @Failure      400  {object}  validation.ErrorResponse  "Invalid ID"
// @Failure      404  {object}  validation.ErrorResponse  "Author not found"
// @Failure      500  {object}  validation.ErrorResponse  "Internal server error"
// @Router       /authors/{id} [get]
func (h *AuthorHandler) GetAuthorByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"AUTHOR_INVALID_ID",
			"invalid author id",
		)
		return
	}

	var author model.Author
	if err := h.db.Preload("Books").First(&author, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(c, http.StatusNotFound,
				"AUTHOR_NOT_FOUND",
				"author not found",
			)
			return
		}

		writeError(c, http.StatusInternalServerError,
			"AUTHOR_FETCH_FAILED",
			"failed to fetch author",
		)
		return
	}

	c.JSON(http.StatusOK, toAuthorResponse(author))
}

// UpdateAuthor godoc
// @Summary      Update an author
// @Description  Partially update an existing author
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        id       path      string               true  "Author ID (UUID)"
// @Param        payload  body      UpdateAuthorRequest  true  "Author fields to update"
// @Success      200      {object}  AuthorResponse
// @Failure      400      {object}  validation.ErrorResponse  "Invalid ID or validation error"
// @Failure      404      {object}  validation.ErrorResponse  "Author not found"
// @Failure      500      {object}  validation.ErrorResponse  "Internal server error"
// @Router       /authors/{id} [patch]
func (h *AuthorHandler) UpdateAuthor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"AUTHOR_INVALID_ID",
			"invalid author id",
		)
		return
	}

	var req UpdateAuthorRequest
	if !validation.BindAndValidateJSON(c, &req) {
		return
	}

	var author model.Author
	if err := h.db.Preload("Books").First(&author, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(c, http.StatusNotFound,
				"AUTHOR_NOT_FOUND",
				"author not found",
			)
			return
		}

		writeError(c, http.StatusInternalServerError,
			"AUTHOR_FETCH_FAILED",
			"failed to fetch author",
		)
		return
	}

	if req.Name != nil {
		author.Name = *req.Name
	}
	if req.Bio != nil {
		author.Bio = *req.Bio
	}

	if err := h.db.Save(&author).Error; err != nil {
		writeError(c, http.StatusInternalServerError,
			"AUTHOR_UPDATE_FAILED",
			"failed to update author",
		)
		return
	}

	c.JSON(http.StatusOK, toAuthorResponse(author))
}

// DeleteAuthor godoc
// @Summary      Delete an author
// @Description  Delete an author by ID
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        id   path      string                    true  "Author ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  validation.ErrorResponse  "Invalid ID"
// @Failure      404  {object}  validation.ErrorResponse  "Author not found"
// @Failure      500  {object}  validation.ErrorResponse  "Internal server error"
// @Router       /authors/{id} [delete]
func (h *AuthorHandler) DeleteAuthor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(c, http.StatusBadRequest,
			"AUTHOR_INVALID_ID",
			"invalid author id",
		)
		return
	}

	result := h.db.Delete(&model.Author{}, "id = ?", id)
	if result.Error != nil {
		writeError(c, http.StatusInternalServerError,
			"AUTHOR_DELETE_FAILED",
			"failed to delete author",
		)
		return
	}

	if result.RowsAffected == 0 {
		writeError(c, http.StatusNotFound,
			"AUTHOR_NOT_FOUND",
			"author not found",
		)
		return
	}

	c.Status(http.StatusNoContent)
}
