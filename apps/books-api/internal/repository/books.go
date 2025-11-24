package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"gorm.io/gorm"
)

type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Book, error)
	List(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormBookRepository struct {
	db *gorm.DB
}

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

func (r *GormBookRepository) List(ctx context.Context) ([]model.Book, error) {
	var books []model.Book
	if err := r.db.WithContext(ctx).
		Preload("Author").
		Find(&books).Error; err != nil {

		return nil, err
	}
	return books, nil
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
