// Package lending owns physical copies, borrowing and private conversations.
package lending

import (
	"time"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"gorm.io/gorm"
)

type Copy struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	BookID       uuid.UUID  `json:"bookId" gorm:"type:uuid;not null;index"`
	Book         model.Book `json:"book" gorm:"constraint:OnDelete:RESTRICT"`
	OwnerID      string     `json:"ownerId" gorm:"not null;index"`
	Condition    string     `json:"condition"`
	Notes        string     `json:"notes"`
	Visible      bool       `json:"visible"`
	Lendable     bool       `json:"lendable"`
	Archived     bool       `json:"archived"`
	Availability string     `json:"availability" gorm:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Request struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CopyID          uuid.UUID  `json:"copyId" gorm:"type:uuid;not null;index"`
	Copy            Copy       `json:"copy" gorm:"constraint:OnDelete:RESTRICT"`
	OwnerID         string     `json:"ownerId" gorm:"not null;index"`
	BorrowerID      string     `json:"borrowerId" gorm:"not null;index;check:no_self_borrow,borrower_id <> owner_id"`
	Status          string     `json:"status" gorm:"not null;index;check:request_status,status IN ('pending','accepted','borrowed','return_pending','returned','declined','cancelled')"`
	Message         string     `json:"message"`
	DueAt           *time.Time `json:"dueAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	AcceptedAt      *time.Time `json:"acceptedAt"`
	BorrowedAt      *time.Time `json:"borrowedAt"`
	ReturnPendingAt *time.Time `json:"returnPendingAt"`
	ClosedAt        *time.Time `json:"closedAt"`
}

type Conversation struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Key         string     `json:"-" gorm:"uniqueIndex;not null"`
	UserA       string     `json:"userA" gorm:"not null;index"`
	UserB       string     `json:"userB" gorm:"not null;index"`
	RequestID   *uuid.UUID `json:"requestId" gorm:"type:uuid;index"`
	Request     *Request   `json:"request,omitempty" gorm:"constraint:OnDelete:RESTRICT"`
	Status      string     `json:"status"`
	Initiator   string     `json:"initiator"`
	AReadAt     time.Time  `json:"-"`
	BReadAt     time.Time  `json:"-"`
	Unread      int64      `json:"unread" gorm:"-"`
	PeerReadAt  time.Time  `json:"peerReadAt" gorm:"-"`
	LastMessage *Message   `json:"lastMessage,omitempty" gorm:"-"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Message struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	ConversationID uuid.UUID    `json:"conversationId" gorm:"type:uuid;not null;index"`
	Conversation   Conversation `json:"-" gorm:"constraint:OnDelete:RESTRICT"`
	SenderID       string       `json:"senderId" gorm:"not null"`
	Body           string       `json:"body"`
	CreatedAt      time.Time    `json:"createdAt" gorm:"index"`
}

type Block struct {
	OwnerID   string    `json:"ownerId" gorm:"primaryKey"`
	UserID    string    `json:"userId" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
}

type SchemaMigration struct {
	Version   int `gorm:"primaryKey"`
	AppliedAt time.Time
}

// Migrate is shared by production and tests. Existing catalog rows are never
// assigned an owner. The partial indexes enforce lending invariants even when
// competing requests arrive through separate database connections.
func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(782374812)").Error; err != nil {
				return err
			}
		}
		if err := tx.AutoMigrate(&model.Author{}, &model.Book{}, &Copy{}, &Request{}, &Conversation{}, &Message{}, &Block{}, &SchemaMigration{}); err != nil {
			return err
		}
		var applied int64
		if err := tx.Model(&SchemaMigration{}).Where("version = 1").Count(&applied).Error; err != nil {
			return err
		}
		if applied > 0 {
			return nil
		}
		for _, sql := range []string{
			`CREATE UNIQUE INDEX IF NOT EXISTS one_occupied_copy ON requests(copy_id) WHERE status IN ('accepted','borrowed','return_pending')`,
			`CREATE UNIQUE INDEX IF NOT EXISTS one_open_request ON requests(copy_id, borrower_id) WHERE status IN ('pending','accepted','borrowed','return_pending')`,
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return tx.Create(&SchemaMigration{Version: 1, AppliedAt: time.Now().UTC()}).Error
	})
}
