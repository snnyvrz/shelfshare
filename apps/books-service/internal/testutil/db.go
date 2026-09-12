package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:testdb_" + uuid.New().String() + "?mode=memory&cache=shared"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&model.Author{}, &model.Book{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func NewErrorDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:errdb_" + uuid.New().String() + "?mode=memory&cache=shared"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to error test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func SeedAuthor(t *testing.T, db *gorm.DB, name string) model.Author {
	t.Helper()

	author := model.Author{
		Name: name,
	}

	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("failed to seed author %q: %v", name, err)
	}

	return author
}

func SeedBook(
	t *testing.T,
	db *gorm.DB,
	author model.Author,
	title, description string,
	publishedAt *time.Time,
) model.Book {
	t.Helper()

	now := time.Now()

	book := model.Book{
		ID:          uuid.New(),
		Title:       title,
		AuthorID:    author.ID,
		Description: description,
		PublishedAt: publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("failed to seed book %q: %v", title, err)
	}

	return book
}
