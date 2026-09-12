//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/handler"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	DBHost := os.Getenv("POSTGRES_HOST")
	DBPort := os.Getenv("POSTGRES_PORT")
	DBUser := os.Getenv("POSTGRES_USER")
	DBPass := os.Getenv("POSTGRES_PASSWORD")
	DBName := os.Getenv("POSTGRES_DB")
	DBSSLMode := "disable"
	TZ := os.Getenv("TZ")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		DBHost,
		DBUser,
		DBPass,
		DBName,
		DBPort,
		DBSSLMode,
		TZ,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	testDB = db

	if err := db.AutoMigrate(&model.Author{}, &model.Book{}); err != nil {
		panic("failed to migrate: " + err.Error())
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	authorRepo := repository.NewAuthorRepository(db)
	bookRepo := repository.NewGormBookRepository(db)

	authorHandler := handler.NewAuthorHandler(authorRepo)
	bookHandler := handler.NewBookHandler(bookRepo)

	api := r.Group("/api")
	{
		authorHandler.RegisterRoutes(api.Group(""))
		bookHandler.RegisterRoutes(api.Group(""))
	}

	testRouter = r

	code := m.Run()
	os.Exit(code)
}

func resetDB(t *testing.T) {
	t.Helper()
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB failed: %v", err)
	}
	_, err = sqlDB.Exec("TRUNCATE TABLE books, authors RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(testRouter)
}

func createTestAuthor(t *testing.T, client *http.Client, baseURL string, name, bio string) string {
	t.Helper()

	reqBody := map[string]any{
		"name": name,
		"bio":  bio,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal author request: %v", err)
	}

	resp, err := client.Post(baseURL+"/api/authors", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 when creating author, got %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode author response: %v", err)
	}

	id := body.Data.ID
	if id == "" {
		t.Fatalf("expected author id in response, got %#v", body.Data.ID)
	}

	return id
}

func createTestBook(t *testing.T, client *http.Client, baseURL, authorID, title, desc string) string {
	t.Helper()

	reqBody := map[string]any{
		"title":       title,
		"author_id":   authorID,
		"description": desc,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal book request: %v", err)
	}

	resp, err := client.Post(baseURL+"/api/books", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 when creating book, got %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode book response: %v", err)
	}

	id := body.Data.ID
	if id == "" {
		t.Fatalf("expected book id in response, got %#v", body.Data.ID)
	}

	return id
}

func TestCreateBookAndFetchIt_BackendIntegration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()

	client := srv.Client()

	authorReq := map[string]string{
		"name": "Robert C. Martin",
		"bio":  "Uncle Bob",
	}
	body, _ := json.Marshal(authorReq)
	resp, err := client.Post(srv.URL+"/api/authors", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	type AuthorCreateResponse struct {
		Data map[string]any `json:"data"`
	}

	var createdAuthor AuthorCreateResponse
	_ = json.NewDecoder(resp.Body).Decode(&createdAuthor)
	resp.Body.Close()

	authorID, _ := createdAuthor.Data["id"].(string)
	bookReq := map[string]any{
		"title":       "Clean Code",
		"author_id":   authorID,
		"description": "A Handbook of Agile Software Craftsmanship",
	}
	body, _ = json.Marshal(bookReq)
	resp, err = client.Post(srv.URL+"/api/books", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	type BookCreateResponse struct {
		Data map[string]any `json:"data"`
	}

	var createdBook BookCreateResponse
	_ = json.NewDecoder(resp.Body).Decode(&createdBook)
	resp.Body.Close()

	bookID, _ := createdBook.Data["id"].(string)
	resp, err = client.Get(srv.URL + "/api/books/" + bookID)
	if err != nil {
		t.Fatalf("failed to get book: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var fetched map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&fetched)

	if fetched["data"].(map[string]any)["title"] != "Clean Code" {
		t.Errorf("expected title=Clean Code, got %v", fetched["data"].(map[string]any)["title"])
	}

	author, ok := fetched["data"].(map[string]any)["author"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'author' object, got %T (%v)", fetched["data"].(map[string]any)["author"], fetched["data"].(map[string]any)["author"])
	}
	if author["id"] != authorID {
		t.Errorf("expected author.id=%s, got %v", authorID, author["id"])
	}
}

func TestCreateAuthor_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()

	client := srv.Client()

	t.Run("success", func(t *testing.T) {
		reqBody := map[string]any{
			"name": "Robert C. Martin",
			"bio":  "Uncle Bob",
		}
		b, _ := json.Marshal(reqBody)

		resp, err := client.Post(srv.URL+"/api/authors", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}

		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if body["data"].(map[string]any)["name"] != "Robert C. Martin" {
			t.Errorf("expected name %q, got %v", "Robert C. Martin", body["data"].(map[string]any)["name"])
		}
		if body["data"].(map[string]any)["bio"] != "Uncle Bob" {
			t.Errorf("expected bio %q, got %v", "Uncle Bob", body["data"].(map[string]any)["bio"])
		}
		if body["data"].(map[string]any)["id"] == "" {
			t.Errorf("expected non-empty id, got %v", body["data"].(map[string]any)["id"])
		}
	})

	t.Run("missing_name", func(t *testing.T) {
		reqBody := map[string]any{
			"bio": "Nameless author",
		}
		b, _ := json.Marshal(reqBody)

		resp, err := client.Post(srv.URL+"/api/authors", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Fatalf("expected 4xx, got %d", resp.StatusCode)
		}
	})

	t.Run("empty_body", func(t *testing.T) {
		resp, err := client.Post(srv.URL+"/api/authors", "application/json", bytes.NewReader([]byte(`{}`)))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Fatalf("expected 4xx, got %d", resp.StatusCode)
		}
	})
}

func TestGetAuthor_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()

	client := srv.Client()

	authorID := createTestAuthor(t, client, srv.URL, "Kent Beck", "TDD guy")

	t.Run("found", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/api/authors/" + authorID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		if body["data"].(map[string]any)["id"] != authorID {
			t.Errorf("expected id %s, got %v", authorID, body["data"].(map[string]any)["id"])
		}
		if body["data"].(map[string]any)["name"] != "Kent Beck" {
			t.Errorf("expected name Kent Beck, got %v", body["data"].(map[string]any)["name"])
		}
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/api/authors/not-a-uuid")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Fatalf("expected 4xx for invalid uuid, got %d", resp.StatusCode)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		randomID := uuid.NewString()
		resp, err := client.Get(srv.URL + "/api/authors/" + randomID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound && (resp.StatusCode < 400 || resp.StatusCode >= 500) {
			t.Fatalf("expected 404 or other 4xx for missing author, got %d", resp.StatusCode)
		}
	})
}

func TestCreateBook_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()

	client := srv.Client()

	authorID := createTestAuthor(t, client, srv.URL, "Robert C. Martin", "Uncle Bob")

	type testCase struct {
		name            string
		payload         map[string]any
		expect2xx       bool
		expectClientErr bool
	}

	cases := []testCase{
		{
			name: "valid_book",
			payload: map[string]any{
				"title":       "Clean Code",
				"author_id":   authorID,
				"description": "A handbook of agile software craftsmanship",
			},
			expect2xx: true,
		},
		{
			name: "missing_title",
			payload: map[string]any{
				"author_id": authorID,
			},
			expectClientErr: true,
		},
		{
			name: "missing_author_id",
			payload: map[string]any{
				"title": "Orphan Book",
			},
			expectClientErr: true,
		},
		{
			name: "invalid_author_id_format",
			payload: map[string]any{
				"title":     "Bad Author",
				"author_id": "not-a-uuid",
			},
			expectClientErr: true,
		},
		{
			name: "nonexistent_author",
			payload: map[string]any{
				"title":     "Ghost Author Book",
				"author_id": uuid.NewString(),
			},
			expectClientErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.payload)
			resp, err := client.Post(srv.URL+"/api/books", "application/json", bytes.NewReader(b))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if tc.expect2xx {
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					t.Fatalf("expected 2xx, got %d", resp.StatusCode)
				}
			}

			if tc.expectClientErr {
				if resp.StatusCode < 400 || resp.StatusCode >= 500 {
					t.Fatalf("expected 4xx, got %d", resp.StatusCode)
				}
			}
		})
	}
}

func TestCreateBookAndFetchIt_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()
	client := srv.Client()

	authorID := createTestAuthor(t, client, srv.URL, "Robert C. Martin", "Uncle Bob")
	bookID := createTestBook(t, client, srv.URL, authorID, "Clean Code", "A classic")

	resp, err := client.Get(srv.URL + "/api/books/" + bookID)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if body["data"].(map[string]any)["id"] != bookID {
		t.Errorf("expected id %s, got %v", bookID, body["data"].(map[string]any)["id"])
	}
	if body["data"].(map[string]any)["title"] != "Clean Code" {
		t.Errorf("expected title Clean Code, got %v", body["data"].(map[string]any)["title"])
	}

	if authorVal, ok := body["data"].(map[string]any)["author"]; ok && authorVal != nil {
		if authorObj, ok := authorVal.(map[string]any); ok {
			if authorObj["id"] != authorID {
				t.Errorf("expected author.id=%s, got %v", authorID, authorObj["id"])
			}
		}
	}
}

func TestGetBook_Errors_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()
	client := srv.Client()

	t.Run("invalid_uuid", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/api/books/not-a-uuid")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Fatalf("expected 4xx, got %d", resp.StatusCode)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/api/books/" + uuid.NewString())
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound && (resp.StatusCode < 400 || resp.StatusCode >= 500) {
			t.Fatalf("expected 404 or other 4xx, got %d", resp.StatusCode)
		}
	})
}

func TestGetAuthorWithBooks_Integration(t *testing.T) {
	resetDB(t)

	srv := newTestServer()
	defer srv.Close()
	client := srv.Client()

	authorID := createTestAuthor(t, client, srv.URL, "Martin Fowler", "Refactoring")
	book1 := createTestBook(t, client, srv.URL, authorID, "Refactoring", "Improving the design of existing code")
	book2 := createTestBook(t, client, srv.URL, authorID, "Patterns of Enterprise Application Architecture", "PoEAA")

	resp, err := client.Get(srv.URL + "/api/authors/" + authorID)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	type BookResponse struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	type AuthorWithBooksResponse struct {
		ID    string         `json:"id"`
		Name  string         `json:"name"`
		Bio   string         `json:"bio"`
		Books []BookResponse `json:"books"`
	}

	var body struct {
		Data AuthorWithBooksResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if body.Data.ID != authorID {
		t.Errorf("expected author id %s, got %s", authorID, body.Data.ID)
	}
	if len(body.Data.Books) != 2 {
		t.Fatalf("expected 2 books, got %d", len(body.Data.Books))
	}

	ids := map[string]bool{
		book1: false,
		book2: false,
	}

	for _, b := range body.Data.Books {
		if _, ok := ids[b.ID]; ok {
			ids[b.ID] = true
		}
	}
	for id, seen := range ids {
		if !seen {
			t.Errorf("expected book id %s in response", id)
		}
	}
}
