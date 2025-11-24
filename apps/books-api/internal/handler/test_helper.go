package handler

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:testdb_" + uuid.New().String() + "?mode=memory&cache=shared"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

func setupErrorDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:errdb_" + uuid.New().String() + "?mode=memory&cache=shared"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	bookRepo := repository.NewGormBookRepository(db)
	bh := NewBookHandler(bookRepo)
	bh.RegisterRoutes(r.Group(""))
	ah := NewAuthorHandler(db)
	ah.RegisterRoutes(r.Group(""))

	return r
}

func seedAuthor(t *testing.T, db *gorm.DB, name string) model.Author {
	t.Helper()

	author := model.Author{
		Name: name,
	}

	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("failed to seed author %q: %v", name, err)
	}

	return author
}

func seedBook(t *testing.T, db *gorm.DB, author model.Author, title, description string, publishedAt *time.Time) model.Book {
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
