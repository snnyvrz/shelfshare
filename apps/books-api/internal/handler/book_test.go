package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
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

	if err := db.AutoMigrate(&model.Book{}); err != nil {
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

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewBookHandler(db)

	h.RegisterRoutes(r.Group(""))

	return r
}

func TestCreateBook_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	body := CreateBookRequest{
		Title:       "Clean Code",
		Author:      "Robert C. Martin",
		Description: "A handbook of Agile software craftsmanship",
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID == uuid.Nil {
		t.Errorf("expected non-empty ID")
	}
	if resp.Title != body.Title {
		t.Errorf("expected title %q, got %q", body.Title, resp.Title)
	}
	if resp.Author != body.Author {
		t.Errorf("expected author %q, got %q", body.Author, resp.Author)
	}
	if resp.PublishedAt != nil {
		t.Errorf("expected PublishedAt to be nil when not provided")
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", resp.ID).Error; err != nil {
		t.Fatalf("expected book in db, got error: %v", err)
	}
}

func TestCreateBook_SuccessWithPublishedAt(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	payload := map[string]any{
		"title":        "DDD",
		"author":       "Eric Evans",
		"description":  "Blue book",
		"published_at": "2003-08-30",
	}

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.PublishedAt == nil {
		t.Fatalf("expected PublishedAt to be set")
	}

	if resp.PublishedAt.Time.IsZero() {
		t.Errorf("expected non-zero PublishedAt time")
	}
}

func TestCreateBook_ValidationError_MissingTitle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	payload := map[string]any{
		"author": "Some Author",
	}

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestListBooks_Empty(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp []BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp))
	}
}

func TestListBooks_WithData(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	now := time.Now()
	book1 := model.Book{
		ID:          uuid.New(),
		Title:       "Book 1",
		Author:      "Author 1",
		Description: "Desc 1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	book2 := model.Book{
		ID:          uuid.New(),
		Title:       "Book 2",
		Author:      "Author 2",
		Description: "Desc 2",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("failed to seed book1: %v", err)
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("failed to seed book2: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp []BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 books, got %d", len(resp))
	}

	found1 := false
	found2 := false
	for _, b := range resp {
		if b.ID == book1.ID {
			found1 = true
		}
		if b.ID == book2.ID {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("expected both seeded books to be present, got %+v", resp)
	}
}

func TestGetBookByID_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	now := time.Now()
	book := model.Book{
		ID:          uuid.New(),
		Title:       "DDD",
		Author:      "Evans",
		Description: "Blue book",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("failed to seed db: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/books/"+book.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != book.ID {
		t.Errorf("expected id %s, got %s", book.ID, resp.ID)
	}
	if resp.Title != book.Title {
		t.Errorf("expected title %q, got %q", book.Title, resp.Title)
	}
}

func TestGetBookByID_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/books/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "INVALID_BOOK_ID" {
		t.Errorf("expected error code INVALID_BOOK_ID, got %q", resp.Code)
	}
}

func TestGetBookByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/books/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_NOT_FOUND" {
		t.Errorf("expected error code BOOK_NOT_FOUND, got %q", resp.Code)
	}
}

func TestUpdateBook_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	now := time.Now()
	book := model.Book{
		ID:          uuid.New(),
		Title:       "Old Title",
		Author:      "Old Author",
		Description: "Old Desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("failed to seed db: %v", err)
	}

	payload := map[string]any{
		"title":       "New Title",
		"description": "New Desc",
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPatch, "/books/"+book.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp BookResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Title != "New Title" {
		t.Errorf("expected updated title, got %q", resp.Title)
	}
	if resp.Description != "New Desc" {
		t.Errorf("expected updated description, got %q", resp.Description)
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", book.ID).Error; err != nil {
		t.Fatalf("expected book in db, got: %v", err)
	}
	if stored.Title != "New Title" || stored.Description != "New Desc" {
		t.Errorf("db not updated correctly: %+v", stored)
	}
}

func TestUpdateBook_NoFieldsToUpdate(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	now := time.Now()
	book := model.Book{
		ID:          uuid.New(),
		Title:       "Title",
		Author:      "Author",
		Description: "Desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("failed to seed db: %v", err)
	}

	b, _ := json.Marshal(map[string]any{})

	req, _ := http.NewRequest(http.MethodPatch, "/books/"+book.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "NO_FIELDS_TO_UPDATE" {
		t.Errorf("expected error code NO_FIELDS_TO_UPDATE, got %q", resp.Code)
	}
}

func TestDeleteBook_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	now := time.Now()
	book := model.Book{
		ID:          uuid.New(),
		Title:       "To delete",
		Author:      "Author",
		Description: "Desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("failed to seed db: %v", err)
	}

	req, _ := http.NewRequest(http.MethodDelete, "/books/"+book.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d, body=%s", w.Code, w.Body.String())
	}

	var count int64
	if err := db.Model(&model.Book{}).Where("id = ?", book.ID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count books: %v", err)
	}
	if count != 0 {
		t.Errorf("expected book to be deleted, still %d records", count)
	}
}

func TestDeleteBook_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodDelete, "/books/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "INVALID_BOOK_ID" {
		t.Errorf("expected error code INVALID_BOOK_ID, got %q", resp.Code)
	}
}

func TestDeleteBook_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodDelete, "/books/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_NOT_FOUND" {
		t.Errorf("expected error code BOOK_NOT_FOUND, got %q", resp.Code)
	}
}
