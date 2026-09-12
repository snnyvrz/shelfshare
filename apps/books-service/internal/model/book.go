package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Book struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title       string    `gorm:"not null;index"`
	AuthorID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Author      Author    `gorm:"foreignKey:AuthorID"`
	Description string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (b *Book) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
