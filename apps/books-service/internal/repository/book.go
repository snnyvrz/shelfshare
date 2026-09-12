package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"gorm.io/gorm"
)

type BookListParams struct {
	Page      int
	PageSize  int
	Sort      string
	Query     string
	AuthorID  *uuid.UUID
	PubAfter  *time.Time
	PubBefore *time.Time
}

type BookListResult struct {
	Books []model.Book
	Total int64
}

type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Book, error)
	List(ctx context.Context, params BookListParams) (BookListResult, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormBookRepository struct {
	db *gorm.DB
}

var ErrAuthorNotFound = errors.New("author not found")

func NewGormBookRepository(db *gorm.DB) *GormBookRepository {
	return &GormBookRepository{db: db}
}

func (r *GormBookRepository) Create(ctx context.Context, book *model.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

func (r *GormBookRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	var book model.Book
	if err := r.db.WithContext(ctx).
		Preload("Author").
		First(&book, "id = ?", id).Error; err != nil {

		return nil, err
	}
	return &book, nil
}

func (r *GormBookRepository) List(ctx context.Context, params BookListParams) (BookListResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}

	db := r.db.WithContext(ctx).Model(&model.Book{}).Preload("Author")

	if params.AuthorID != nil {
		db = db.Where("author_id = ?", *params.AuthorID)
	}

	if params.PubAfter != nil {
		db = db.Where("published_at >= ?", *params.PubAfter)
	}

	if params.PubBefore != nil {
		db = db.Where("published_at <= ?", *params.PubBefore)
	}

	dialect := r.db.Dialector.Name()
	if params.Query != "" {
		like := "%" + params.Query + "%"
		if dialect == "postgres" {
			db = db.Where(
				"title ILIKE ? OR description ILIKE ?",
				like, like,
			)
		} else {
			q := strings.ToLower(params.Query)
			like := "%" + q + "%"
			db = db.Where(
				"LOWER(title) LIKE ? OR LOWER(description) LIKE ?",
				like, like,
			)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return BookListResult{}, err
	}

	switch params.Sort {
	case "title_asc":
		db = db.Order("title ASC")
	case "title_desc":
		db = db.Order("title DESC")
	case "published_at_asc":
		db = db.Order("published_at ASC NULLS LAST")
	case "published_at_desc":
		db = db.Order("published_at DESC NULLS LAST")
	case "created_at_asc":
		db = db.Order("created_at ASC")
	case "created_at_desc", "":
		fallthrough
	default:
		db = db.Order("created_at DESC")
	}

	offset := (params.Page - 1) * params.PageSize

	var books []model.Book
	if err := db.
		Limit(params.PageSize).
		Offset(offset).
		Find(&books).Error; err != nil {

		return BookListResult{}, err
	}

	return BookListResult{
		Books: books,
		Total: total,
	}, nil
}

func (r *GormBookRepository) Update(ctx context.Context, book *model.Book) error {
	return r.db.WithContext(ctx).
		Model(&model.Book{}).
		Where("id = ?", book.ID).
		Updates(map[string]any{
			"title":        book.Title,
			"description":  book.Description,
			"author_id":    book.AuthorID,
			"published_at": book.PublishedAt,
		}).Error
}

func (r *GormBookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Book{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
