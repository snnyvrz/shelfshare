package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
	"gorm.io/gorm"
)

func TestCreateAuthor_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	body := CreateAuthorRequest{
		Name: "Martin Fowler",
		Bio:  "Author of many software books",
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/authors", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp AuthorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID == uuid.Nil {
		t.Errorf("expected non-empty ID")
	}
	if resp.Name != body.Name {
		t.Errorf("expected name %q, got %q", body.Name, resp.Name)
	}
	if resp.Bio != body.Bio {
		t.Errorf("expected bio %q, got %q", body.Bio, resp.Bio)
	}

	var stored model.Author
	if err := db.First(&stored, "id = ?", resp.ID).Error; err != nil {
		t.Fatalf("expected author in db, got error: %v", err)
	}

	if stored.Name != body.Name {
		t.Errorf("expected stored name %q, got %q", body.Name, stored.Name)
	}
	if stored.Bio != body.Bio {
		t.Errorf("expected stored bio %q, got %q", body.Bio, stored.Bio)
	}
}

func TestCreateAuthor_ValidationError_MissingName(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	payload := map[string]any{
		"bio": "Some bio",
	}

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/authors", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCreateAuthor_InternalError_Returns500(t *testing.T) {
	db := setupErrorDB(t)
	router := setupRouter(db)

	body := CreateAuthorRequest{
		Name: "Error Author",
		Bio:  "Should fail",
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/authors", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != "AUTHOR_CREATE_FAILED" {
		t.Errorf("expected error code AUTHOR_CREATE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to create author" {
		t.Errorf("expected message %q, got %q", "failed to create author", resp.Message)
	}
}

func TestListAuthors_Empty(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/authors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp []AuthorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp))
	}
}

func TestListAuthors_WithData(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	author1 := seedAuthor(t, db, "Author 1")
	author2 := seedAuthor(t, db, "Author 2")

	req, _ := http.NewRequest(http.MethodGet, "/authors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp []AuthorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(resp))
	}

	found1 := false
	found2 := false

	for _, a := range resp {
		switch a.ID {
		case author1.ID:
			found1 = true
			if a.Name != author1.Name {
				t.Errorf("expected author1 name %q, got %q", author1.Name, a.Name)
			}
		case author2.ID:
			found2 = true
			if a.Name != author2.Name {
				t.Errorf("expected author2 name %q, got %q", author2.Name, a.Name)
			}
		}
	}

	if !found1 || !found2 {
		t.Errorf("expected both seeded authors to be present, got %+v", resp)
	}
}

func TestListAuthors_InternalError_Returns500(t *testing.T) {
	db := setupErrorDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/authors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_LIST_FAILED" {
		t.Errorf("expected error code AUTHOR_LIST_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to list authors" {
		t.Errorf("expected message %q, got %q", "failed to list authors", resp.Message)
	}
}

func TestGetAuthorByID_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	author := seedAuthor(t, db, "Evans")

	req, _ := http.NewRequest(http.MethodGet, "/authors/"+author.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp AuthorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != author.ID {
		t.Errorf("expected id %s, got %s", author.ID, resp.ID)
	}
	if resp.Name != author.Name {
		t.Errorf("expected name %q, got %q", author.Name, resp.Name)
	}
}

func TestGetAuthorByID_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/authors/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_INVALID_ID" {
		t.Errorf("expected error code AUTHOR_INVALID_ID, got %q", resp.Code)
	}
}

func TestGetAuthorByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/authors/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_NOT_FOUND" {
		t.Errorf("expected error code AUTHOR_NOT_FOUND, got %q", resp.Code)
	}
}

func TestGetAuthorByID_InternalError_Returns500(t *testing.T) {
	db := setupErrorDB(t)
	router := setupRouter(db)

	id := "550e8400-e29b-41d4-a716-446655440000"
	req, _ := http.NewRequest(http.MethodGet, "/authors/"+id, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_FETCH_FAILED" {
		t.Errorf("expected error code AUTHOR_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch author" {
		t.Errorf("expected message %q, got %q", "failed to fetch author", resp.Message)
	}
}

func TestUpdateAuthor_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	author := seedAuthor(t, db, "Old Name")

	if err := db.Model(&author).Update("bio", "Old Bio").Error; err != nil {
		t.Fatalf("failed to update seed author bio: %v", err)
	}

	payload := map[string]any{
		"name": "New Name",
		"bio":  "New Bio",
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPatch, "/authors/"+author.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp AuthorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "New Name" {
		t.Errorf("expected updated name, got %q", resp.Name)
	}
	if resp.Bio != "New Bio" {
		t.Errorf("expected updated bio, got %q", resp.Bio)
	}

	var stored model.Author
	if err := db.First(&stored, "id = ?", author.ID).Error; err != nil {
		t.Fatalf("expected author in db, got: %v", err)
	}
	if stored.Name != "New Name" || stored.Bio != "New Bio" {
		t.Errorf("db not updated correctly (name/bio): %+v", stored)
	}
}

func TestUpdateAuthor_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	payload := map[string]any{
		"name": "Doesn't matter",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPatch, "/authors/not-a-uuid", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_INVALID_ID" {
		t.Errorf("expected error code AUTHOR_INVALID_ID, got %q", resp.Code)
	}
}

func TestUpdateAuthor_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	nonExistentID := uuid.New().String()

	payload := map[string]any{
		"name": "New Name",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/authors/"+nonExistentID,
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

	if resp.Code != "AUTHOR_NOT_FOUND" {
		t.Errorf("expected error code AUTHOR_NOT_FOUND, got %q", resp.Code)
	}
	if resp.Message != "author not found" {
		t.Errorf("expected message %q, got %q", "author not found", resp.Message)
	}
}

func TestUpdateAuthor_ValidationError_InvalidName(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	author := seedAuthor(t, db, "Author")

	payload := map[string]any{
		"name": "",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/authors/"+author.ID.String(),
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

func TestUpdateAuthor_InternalErrorOnFetch_Returns500(t *testing.T) {
	db := setupErrorDB(t)
	router := setupRouter(db)

	id := "550e8400-e29b-41d4-a716-446655440000"
	payload := map[string]any{
		"name": "Updated name",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPatch, "/authors/"+id, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_FETCH_FAILED" {
		t.Errorf("expected error code AUTHOR_FETCH_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to fetch author" {
		t.Errorf("expected message %q, got %q", "failed to fetch author", resp.Message)
	}
}

func TestUpdateAuthor_InternalErrorOnSave_Returns500(t *testing.T) {
	db := setupTestDB(t)

	author := seedAuthor(t, db, "Original")

	if err := db.Callback().Update().
		Before("gorm:before_update").
		Register("force_author_update_error", func(tx *gorm.DB) {
			if tx.Statement.Table == "authors" {
				tx.AddError(errors.New("forced update error"))
			}
		}); err != nil {
		t.Fatalf("failed to register update callback: %v", err)
	}

	router := setupRouter(db)

	payload := map[string]any{
		"name": "New Name",
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPatch,
		"/authors/"+author.ID.String(),
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

	if resp.Code != "AUTHOR_UPDATE_FAILED" {
		t.Errorf("expected error code AUTHOR_UPDATE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to update author" {
		t.Errorf("expected message %q, got %q", "failed to update author", resp.Message)
	}
}

func TestDeleteAuthor_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	author := seedAuthor(t, db, "Author To Delete")

	req, _ := http.NewRequest(http.MethodDelete, "/authors/"+author.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d, body=%s", w.Code, w.Body.String())
	}

	var count int64
	if err := db.Model(&model.Author{}).Where("id = ?", author.ID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count authors: %v", err)
	}
	if count != 0 {
		t.Errorf("expected author to be deleted, still %d records", count)
	}
}

func TestDeleteAuthor_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodDelete, "/authors/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_INVALID_ID" {
		t.Errorf("expected error code AUTHOR_INVALID_ID, got %q", resp.Code)
	}
}

func TestDeleteAuthor_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(db)

	req, _ := http.NewRequest(http.MethodDelete, "/authors/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_NOT_FOUND" {
		t.Errorf("expected error code AUTHOR_NOT_FOUND, got %q", resp.Code)
	}
}

func TestDeleteAuthor_InternalError_Returns500(t *testing.T) {
	db := setupErrorDB(t)
	router := setupRouter(db)

	id := "550e8400-e29b-41d4-a716-446655440000"
	req, _ := http.NewRequest(http.MethodDelete, "/authors/"+id, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp validation.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != "AUTHOR_DELETE_FAILED" {
		t.Errorf("expected error code AUTHOR_DELETE_FAILED, got %q", resp.Code)
	}
	if resp.Message != "failed to delete author" {
		t.Errorf("expected message %q, got %q", "failed to delete author", resp.Message)
	}
}
