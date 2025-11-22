package repository

import (
	"gorm.io/gorm"

	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db}
}

func (r *BookRepository) Create(book *model.Book) error {
	return r.db.Create(book).Error
}

func (r *BookRepository) FindAll() ([]model.Book, error) {
	var books []model.Book
	err := r.db.Find(&books).Error
	return books, err
}
