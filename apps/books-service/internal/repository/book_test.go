package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
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

func seedBooks(t *testing.T, db *gorm.DB) (model.Author, model.Author) {
	t.Helper()

	author1 := model.Author{
		ID:   uuid.New(),
		Name: "Author One",
		Bio:  "A1",
	}
	author2 := model.Author{
		ID:   uuid.New(),
		Name: "Author Two",
		Bio:  "A2",
	}

	if err := db.Create(&author1).Error; err != nil {
		t.Fatalf("failed to seed author1: %v", err)
	}
	if err := db.Create(&author2).Error; err != nil {
		t.Fatalf("failed to seed author2: %v", err)
	}

	now := time.Now()

	books := []model.Book{
		{
			ID:        uuid.New(),
			Title:     "Clean Code",
			AuthorID:  author1.ID,
			CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Title:     "Clean Architecture",
			AuthorID:  author1.ID,
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Title:     "Domain-Driven Design",
			AuthorID:  author2.ID,
			CreatedAt: now.Add(-1 * time.Hour),
		},
	}

	if err := db.Create(&books).Error; err != nil {
		t.Fatalf("failed to seed books: %v", err)
	}

	return author1, author2
}

func TestGormBookRepository_List_SearchAndSortAndPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormBookRepository(db)

	_, _ = seedBooks(t, db)

	ctx := context.Background()

	params := BookListParams{
		Page:     1,
		PageSize: 10,
		Sort:     "title_asc",
		Query:    "Clean",
	}

	result, err := repo.List(ctx, params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total=2, got %d", result.Total)
	}

	if len(result.Books) != 2 {
		t.Fatalf("expected 2 books, got %d", len(result.Books))
	}

	if result.Books[0].Title != "Clean Architecture" || result.Books[1].Title != "Clean Code" {
		t.Fatalf("unexpected order: got [%s, %s]",
			result.Books[0].Title,
			result.Books[1].Title,
		)
	}
}

func TestGormBookRepository_List_FilterByAuthorAndPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormBookRepository(db)

	_, author2 := seedBooks(t, db)
	ctx := context.Background()

	params := BookListParams{
		Page:     1,
		PageSize: 1,
		AuthorID: &author2.ID,
		Sort:     "created_at_desc",
	}

	result, err := repo.List(ctx, params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total=1 for author2, got %d", result.Total)
	}

	if len(result.Books) != 1 {
		t.Fatalf("expected 1 book on page 1, got %d", len(result.Books))
	}

	if result.Books[0].AuthorID != author2.ID {
		t.Fatalf("expected book author_id=%s, got %s", author2.ID, result.Books[0].AuthorID)
	}
}
