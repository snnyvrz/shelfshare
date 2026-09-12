package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"gorm.io/gorm"
)

type AuthorRepository interface {
	Create(ctx context.Context, author *model.Author) error
	List(ctx context.Context) ([]model.Author, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Author, error)
	Update(ctx context.Context, author *model.Author) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormAuthorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) AuthorRepository {
	return &GormAuthorRepository{db: db}
}

func (r *GormAuthorRepository) Create(ctx context.Context, author *model.Author) error {
	return r.db.WithContext(ctx).Create(author).Error
}

func (r *GormAuthorRepository) List(ctx context.Context) ([]model.Author, error) {
	var authors []model.Author

	if err := r.db.WithContext(ctx).
		Preload("Books").
		Order("created_at DESC").
		Find(&authors).Error; err != nil {

		return nil, err
	}

	return authors, nil
}

func (r *GormAuthorRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	var author model.Author

	if err := r.db.WithContext(ctx).
		Preload("Books").
		First(&author, "id = ?", id).Error; err != nil {

		return nil, err
	}

	return &author, nil
}

func (r *GormAuthorRepository) Update(ctx context.Context, author *model.Author) error {
	return r.db.WithContext(ctx).Save(author).Error
}

func (r *GormAuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Author{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
