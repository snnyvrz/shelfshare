package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Author struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null;index"`
	Bio       string
	Books     []Book `json:"books,omitempty" gorm:"foreignKey:AuthorID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *Author) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
