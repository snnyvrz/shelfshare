package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/repository"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/validation"
	"gorm.io/gorm"
)

func setupTestRouterWithRepos(
	bookRepo repository.BookRepository,
	authorRepo repository.AuthorRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	bh := NewBookHandler(bookRepo)
	bh.RegisterRoutes(r.Group(""))

	ah := NewAuthorHandler(authorRepo)
	ah.RegisterRoutes(r.Group(""))

	return r
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
	bookRepo := repository.NewGormBookRepository(db)
	authorRepo := repository.NewAuthorRepository(db)
	return setupTestRouterWithRepos(bookRepo, authorRepo)
}

func parseIntQuery(c *gin.Context, key string, def int) int {
	if s := c.Query(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return def
}

func parseDateQuery(c *gin.Context, key string) (*time.Time, error) {
	s := c.Query(key)
	if s == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, validation.ErrorResponse{
		Code:    code,
		Message: message,
		Errors:  nil,
	})
}

func toBookResponse(b model.Book) BookResponse {
	var pub *model.Date
	if b.PublishedAt != nil && !b.PublishedAt.IsZero() {
		pub = &model.Date{Time: *b.PublishedAt}
	}

	data := Book{
		ID:    b.ID,
		Title: b.Title,
		Author: AuthorSummary{
			ID:   b.Author.ID,
			Name: b.Author.Name,
			Bio:  b.Author.Bio,
		},
		Description: b.Description,
		PublishedAt: pub,
		CreatedAt:   model.Date{Time: b.CreatedAt},
		UpdatedAt:   model.Date{Time: b.UpdatedAt},
	}

	return BookResponse{
		Data: data,
	}
}

func toBookSummaryResponse(b model.Book) BookSummaryResponse {
	var pub *model.Date
	if b.PublishedAt != nil && !b.PublishedAt.IsZero() {
		pub = &model.Date{Time: *b.PublishedAt}
	}

	data := BookSummary{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		PublishedAt: pub,
		CreatedAt:   model.Date{Time: b.CreatedAt},
		UpdatedAt:   model.Date{Time: b.UpdatedAt},
	}

	return BookSummaryResponse{
		Data: data,
	}
}

func toListBooksResponse(br []Book, page, pageSize int, total int64, totalPages int) ListBooksResponse {
	return ListBooksResponse{
		Data: br,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
