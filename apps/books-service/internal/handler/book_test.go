package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/repository"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/testutil"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/validation"
	"gorm.io/gorm"
)

type fakeBookRepo struct {
	CreateFn   func(ctx context.Context, b *model.Book) error
	ListFn     func(ctx context.Context, params repository.BookListParams) (repository.BookListResult, error)
	FindByIDFn func(ctx context.Context, id uuid.UUID) (*model.Book, error)
	UpdateFn   func(ctx context.Context, b *model.Book) error
	DeleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (f *fakeBookRepo) Create(ctx context.Context, b *model.Book) error {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, b)
	}
	return nil
}

func (f *fakeBookRepo) List(ctx context.Context, params repository.BookListParams) (repository.BookListResult, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, params)
	}
	return repository.BookListResult{}, nil
}

func (f *fakeBookRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	if f.FindByIDFn != nil {
		return f.FindByIDFn(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeBookRepo) Update(ctx context.Context, b *model.Book) error {
	if f.UpdateFn != nil {
		return f.UpdateFn(ctx, b)
	}
	return nil
}

func (f *fakeBookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if f.DeleteFn != nil {
		return f.DeleteFn(ctx, id)
	}
	return nil
}

func setupBookRouterWithRepo(bookRepo repository.BookRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewBookHandler(bookRepo)
	h.RegisterRoutes(r.Group(""))

	return r
}

func TestCreateBook_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Evans")

	body := CreateBookRequest{
		Title:       "Clean Code",
		AuthorID:    author.ID,
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

	if resp.Data.ID == uuid.Nil {
		t.Errorf("expected non-empty ID")
	}
	if resp.Data.Title != body.Title {
		t.Errorf("expected title %q, got %q", body.Title, resp.Data.Title)
	}

	if resp.Data.Author.ID != author.ID {
		t.Errorf("expected author ID %q, got %q", author.ID, resp.Data.Author.ID)
	}
	if resp.Data.Author.Name != author.Name {
		t.Errorf("expected author name %q, got %q", author.Name, resp.Data.Author.Name)
	}

	if resp.Data.PublishedAt != nil {
		t.Errorf("expected PublishedAt to be nil when not provided")
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", resp.Data.ID).Error; err != nil {
		t.Fatalf("expected book in db, got error: %v", err)
	}

	if stored.AuthorID != author.ID {
		t.Errorf("expected stored AuthorID %q, got %q", author.ID, stored.AuthorID)
	}
}

func TestCreateBook_SuccessWithPublishedAt(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Evans")

	payload := map[string]any{
		"title":        "Clean Code",
		"author_id":    author.ID.String(),
		"description":  "A handbook of Agile software craftsmanship",
		"published_at": "2020-01-01",
	}

	b, err := json.Marshal(payload)
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

	if resp.Data.PublishedAt == nil {
		t.Fatalf("expected PublishedAt to be non-nil")
	}
	if got := resp.Data.PublishedAt.Time.Format("2006-01-02"); got != "2020-01-01" {
		t.Errorf("expected PublishedAt 2020-01-01, got %s", got)
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", resp.Data.ID).Error; err != nil {
		t.Fatalf("expected book in db, got error: %v", err)
	}
	if stored.PublishedAt == nil || stored.PublishedAt.Format("2006-01-02") != "2020-01-01" {
		t.Errorf("expected stored PublishedAt 2020-01-01, got %v", stored.PublishedAt)
	}
}

func TestCreateBook_ValidationError_MissingTitle(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

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

func TestCreateBook_InternalError_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		CreateFn: func(ctx context.Context, b *model.Book) error {
			return errors.New("forced create error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	body := CreateBookRequest{
		Title:       "Error book",
		AuthorID:    uuid.New(),
		Description: "Should fail",
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != "BOOK_CREATE_FAILED" {
		t.Errorf("expected error code BOOK_CREATE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to create book" {
		t.Errorf("expected message %q, got %q", "failed to create book", resp.Message)
	}
}

func TestCreateBook_FetchCreatedBookError_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		CreateFn: func(ctx context.Context, b *model.Book) error {
			if b.ID == uuid.Nil {
				b.ID = uuid.New()
			}
			return nil
		},
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Book, error) {
			return nil, errors.New("forced fetch error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	body := CreateBookRequest{
		Title:       "Book Title",
		AuthorID:    uuid.New(),
		Description: "Some description",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != "BOOK_FETCH_FAILED" {
		t.Errorf("expected error code BOOK_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch created book" {
		t.Errorf("expected message %q, got %q", "failed to fetch created book", resp.Message)
	}
}

func TestListBooks_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp ListBooksResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp.Data))
	}
}

func TestListBooks_WithData(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author1 := testutil.SeedAuthor(t, db, "Author 1")
	author2 := testutil.SeedAuthor(t, db, "Author 2")

	book1 := testutil.SeedBook(t, db, author1, "Book 1", "Desc 1", nil)
	book2 := testutil.SeedBook(t, db, author2, "Book 2", "Desc 2", nil)

	req, _ := http.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp ListBooksResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 books, got %d", len(resp.Data))
	}

	found1 := false
	found2 := false

	for _, b := range resp.Data {
		switch b.ID {
		case book1.ID:
			found1 = true
			if b.Author.ID != author1.ID {
				t.Errorf("expected book1 author ID %q, got %q", author1.ID, b.Author.ID)
			}
			if b.Author.Name != author1.Name {
				t.Errorf("expected book1 author name %q, got %q", author1.Name, b.Author.Name)
			}
		case book2.ID:
			found2 = true
			if b.Author.ID != author2.ID {
				t.Errorf("expected book2 author ID %q, got %q", author2.ID, b.Author.ID)
			}
			if b.Author.Name != author2.Name {
				t.Errorf("expected book2 author name %q, got %q", author2.Name, b.Author.Name)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("expected both seeded books to be present, got %+v", resp)
	}
}

func TestListBooks_InternalError_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		ListFn: func(ctx context.Context, params repository.BookListParams) (repository.BookListResult, error) {
			return repository.BookListResult{}, errors.New("forced list error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	req, _ := http.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_LIST_FAILED" {
		t.Errorf("expected error code BOOK_LIST_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch books" {
		t.Errorf("expected message %q, got %q", "failed to fetch books", resp.Message)
	}
}

func TestGetBookByID_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Evans")
	book := testutil.SeedBook(t, db, author, "DDD", "Blue Book", nil)

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

	if resp.Data.ID != book.ID {
		t.Errorf("expected id %s, got %s", book.ID, resp.Data.ID)
	}
	if resp.Data.Title != book.Title {
		t.Errorf("expected title %q, got %q", book.Title, resp.Data.Title)
	}

	if resp.Data.Author.ID != author.ID {
		t.Errorf("expected author id %s, got %s", author.ID, resp.Data.Author.ID)
	}
	if resp.Data.Author.Name != author.Name {
		t.Errorf("expected author name %q, got %q", author.Name, resp.Data.Author.Name)
	}
}

func TestGetBookByID_InvalidUUID(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

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
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

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

func TestGetBookByID_InternalError_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Book, error) {
			return nil, errors.New("forced fetch error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	id := "550e8400-e29b-41d4-a716-446655440000"
	req, _ := http.NewRequest(http.MethodGet, "/books/"+id, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_FETCH_FAILED" {
		t.Errorf("expected error code BOOK_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch book" {
		t.Errorf("expected message %q, got %q", "failed to fetch book", resp.Message)
	}
}

func TestUpdateBook_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	oldAuthor := testutil.SeedAuthor(t, db, "Old Author")
	newAuthor := testutil.SeedAuthor(t, db, "New Author")

	book := testutil.SeedBook(t, db, oldAuthor, "Old Title", "Old Desc", nil)

	payload := map[string]any{
		"title":        "New Title",
		"author_id":    newAuthor.ID.String(),
		"description":  "New Desc",
		"published_at": "2020-01-01",
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

	if resp.Data.Title != "New Title" {
		t.Errorf("expected updated title, got %q", resp.Data.Title)
	}
	if resp.Data.Description != "New Desc" {
		t.Errorf("expected updated description, got %q", resp.Data.Description)
	}
	if resp.Data.Author.ID != newAuthor.ID {
		t.Errorf("expected author ID %s, got %s", newAuthor.ID, resp.Data.Author.ID)
	}
	if resp.Data.Author.Name != newAuthor.Name {
		t.Errorf("expected author name %q, got %q", newAuthor.Name, resp.Data.Author.Name)
	}
	if resp.Data.PublishedAt == nil || resp.Data.PublishedAt.Time.Format("2006-01-02") != "2020-01-01" {
		t.Errorf("expected PublishedAt 2020-01-01, got %+v", resp.Data.PublishedAt)
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", book.ID).Error; err != nil {
		t.Fatalf("expected book in db, got: %v", err)
	}
	if stored.Title != "New Title" || stored.Description != "New Desc" {
		t.Errorf("db not updated correctly (title/description): %+v", stored)
	}
	if stored.AuthorID != newAuthor.ID {
		t.Errorf("expected stored AuthorID %s, got %s", newAuthor.ID, stored.AuthorID)
	}
	if stored.PublishedAt == nil || stored.PublishedAt.Format("2006-01-02") != "2020-01-01" {
		t.Errorf("expected stored PublishedAt 2020-01-01, got %v", stored.PublishedAt)
	}
}

func TestUpdateBook_InvalidUUID(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	payload := map[string]any{
		"title": "Doesn't matter",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPatch, "/books/not-a-uuid", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

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

func TestUpdateBook_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	nonExistentID := uuid.New().String()

	payload := map[string]any{
		"title": "New Title",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/books/"+nonExistentID,
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

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
	if resp.Message != "book not found" {
		t.Errorf("expected message %q, got %q", "book not found", resp.Message)
	}
}

func TestUpdateBook_NoFieldsToUpdate(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Author")
	book := testutil.SeedBook(t, db, author, "Title", "Desc", nil)

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

func TestUpdateBook_ValidationError_InvalidTitle(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Author")
	book := testutil.SeedBook(t, db, author, "Title", "Desc", nil)

	payload := map[string]any{
		"title": "",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/books/"+book.ID.String(),
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code == "" {
		t.Errorf("expected validation error code to be set, got empty string")
	}
}

func TestUpdateBook_ClearPublishedAt_WhenZeroDate(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	now := time.Now()
	pub := now.Add(-24 * time.Hour)

	author := testutil.SeedAuthor(t, db, "Author")
	book := testutil.SeedBook(t, db, author, "Title", "Desc", &pub)

	payload := map[string]any{
		"published_at": "",
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		http.MethodPatch,
		"/books/"+book.ID.String(),
		bytes.NewReader(b),
	)
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

	if resp.Data.PublishedAt != nil {
		t.Errorf("expected PublishedAt to be nil in response, got %v", resp.Data.PublishedAt)
	}

	var stored model.Book
	if err := db.First(&stored, "id = ?", book.ID).Error; err != nil {
		t.Fatalf("failed to fetch updated book: %v", err)
	}

	if stored.PublishedAt != nil {
		t.Errorf("expected stored PublishedAt to be nil, got %v", stored.PublishedAt)
	}
}

func TestUpdateBook_InternalErrorOnFetch_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Book, error) {
			return nil, errors.New("forced fetch error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	id := "550e8400-e29b-41d4-a716-446655440000"
	payload := map[string]any{
		"title": "Updated title",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPatch, "/books/"+id, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_FETCH_FAILED" {
		t.Errorf("expected error code BOOK_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch book" {
		t.Errorf("expected message %q, got %q", "failed to fetch book", resp.Message)
	}
}

func TestUpdateBook_InternalErrorOnSave_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Book, error) {
			return &model.Book{ID: id, Title: "Original"}, nil
		},
		UpdateFn: func(ctx context.Context, b *model.Book) error {
			return errors.New("forced update error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	id := "550e8400-e29b-41d4-a716-446655440000"
	payload := map[string]any{
		"title": "New Title",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/books/"+id,
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != "BOOK_UPDATE_FAILED" {
		t.Errorf("expected error code BOOK_UPDATE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to update book" {
		t.Errorf("expected message %q, got %q", "failed to update book", resp.Message)
	}
}

func TestUpdateBook_InternalErrorOnFetchUpdated_Returns500(t *testing.T) {
	var findCalls int
	bookRepo := &fakeBookRepo{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Book, error) {
			findCalls++
			if findCalls == 1 {
				return &model.Book{ID: id, Title: "Original"}, nil
			}
			return nil, errors.New("forced fetch updated error")
		},
		UpdateFn: func(ctx context.Context, b *model.Book) error {
			return nil
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	id := "550e8400-e29b-41d4-a716-446655440000"
	payload := map[string]any{
		"title": "New Title",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/books/"+id,
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != "BOOK_FETCH_FAILED" {
		t.Errorf("expected error code BOOK_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch updated book" {
		t.Errorf("expected message %q, got %q", "failed to fetch updated book", resp.Message)
	}
}

func TestDeleteBook_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

	author := testutil.SeedAuthor(t, db, "Author")
	book := testutil.SeedBook(t, db, author, "To Delete", "Desc", nil)

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
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

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
	db := testutil.NewTestDB(t)
	router := setupTestRouter(db)

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

func TestDeleteBook_InternalError_Returns500(t *testing.T) {
	bookRepo := &fakeBookRepo{
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("forced delete error")
		},
	}

	router := setupBookRouterWithRepo(bookRepo)

	id := "550e8400-e29b-41d4-a716-446655440000"
	req, _ := http.NewRequest(http.MethodDelete, "/books/"+id, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "BOOK_DELETE_FAILED" {
		t.Errorf("expected error code BOOK_DELETE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to delete book" {
		t.Errorf("expected message %q, got %q", "failed to delete book", resp.Message)
	}
}
