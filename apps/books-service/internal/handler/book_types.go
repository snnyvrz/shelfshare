package handler

import (
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
)

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

type Book struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Author      AuthorSummary `json:"author"`
	Description string        `json:"description"`
	PublishedAt *model.Date   `json:"published_at,omitempty" swaggertype:"string" example:"2025-11-24"`
	CreatedAt   model.Date    `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt   model.Date    `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
}

type BookResponse struct {
	Data Book `json:"data"`
}

type BookSummary struct {
	ID          uuid.UUID   `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	PublishedAt *model.Date `json:"published_at,omitempty" swaggertype:"string" example:"2025-11-24"`
	CreatedAt   model.Date  `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt   model.Date  `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
}

type BookSummaryResponse struct {
	Data BookSummary `json:"data"`
}

type Pagination struct {
	Page       int   `json:"page" binding:"omitempty,min=1"`
	PageSize   int   `json:"page_size" binding:"omitempty,min=1"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListBooksResponse struct {
	Data       []Book     `json:"data"`
	Pagination Pagination `json:"pagination"`
}
