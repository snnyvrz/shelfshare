package handler

import (
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
)

type CreateAuthorRequest struct {
	Name string `json:"name" binding:"required,min=1"`
	Bio  string `json:"bio" binding:"omitempty,max=2000"`
}

type UpdateAuthorRequest struct {
	Name *string `json:"name" binding:"omitempty,min=1"`
	Bio  *string `json:"bio" binding:"omitempty,max=2000"`
}

type Author struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Bio       string        `json:"bio"`
	Books     []BookSummary `json:"books,omitempty"`
	CreatedAt model.Date    `json:"created_at" swaggertype:"string" example:"2025-11-24"`
	UpdatedAt model.Date    `json:"updated_at" swaggertype:"string" example:"2025-11-24"`
}

type AuthorSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Bio  string    `json:"bio"`
}

type AuthorResponse struct {
	Data Author `json:"data"`
}
