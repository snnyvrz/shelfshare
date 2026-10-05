//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/auth"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/lending"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
)

func TestNearbyCopiesRankAndPaginateCompleteOwnerSet(t *testing.T) {
	resetDB(t)
	a := model.Author{Name: "Discovery writer"}
	if err := testDB.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	b := model.Book{Title: "Nearby book", AuthorID: a.ID}
	if err := testDB.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{Email: "discovery@example.com", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "0123456789abcdef01234567", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}).SignedString([]byte(bddSecret))
	if err != nil {
		t.Fatal(err)
	}
	owners := []map[string]any{}
	for i := 0; i < 52; i++ {
		owner := fmt.Sprintf("%024x", i+1)
		rank := 0
		if i == 51 {
			rank = 1
		}
		owners = append(owners, map[string]any{"ownerId": owner, "rank": rank})
		cp := lending.Copy{ID: uuid.New(), BookID: b.ID, OwnerID: owner, Condition: "Good", Visible: true, Lendable: true,
			CreatedAt: time.Date(2026, 1, 1, 0, i, 0, 0, time.UTC)}
		if err := testDB.Create(&cp).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, cp := range []lending.Copy{
		{ID: uuid.New(), BookID: b.ID, OwnerID: fmt.Sprintf("%024x", 1), Condition: "Hidden", Visible: false},
		{ID: uuid.New(), BookID: b.ID, OwnerID: fmt.Sprintf("%024x", 1), Condition: "Archived", Visible: true, Archived: true},
		{ID: uuid.New(), BookID: b.ID, OwnerID: fmt.Sprintf("%024x", 100), Condition: "Opted out", Visible: true},
	} {
		if err := testDB.Create(&cp).Error; err != nil {
			t.Fatal(err)
		}
	}
	status := http.StatusOK
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("JWT not forwarded")
		}
		if r.URL.Path != "/nearby/owners" || r.URL.Query().Get("cityId") != "112931" || r.URL.Query().Get("mode") != "radius" {
			t.Errorf("unexpected upstream query: %s", r.URL)
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": owners})
	}))
	defer upstream.Close()
	t.Setenv("AUTH_API_URL", upstream.URL)
	router := newLendingRouter(&lending.Service{DB: testDB})
	call := func(page int, authorized bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/nearby/copies?mode=radius&cityId=112931&radiusKm=50&page=%d", page), nil)
		if authorized {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	if rec := call(1, false); rec.Code != 401 {
		t.Fatalf("unauthenticated: %d", rec.Code)
	}
	var pages [2]struct {
		Data  []lending.Copy `json:"data"`
		Total int            `json:"total"`
	}
	for i := range pages {
		rec := call(i+1, true)
		if rec.Code != 200 {
			t.Fatalf("page %d: %d %s", i+1, rec.Code, rec.Body)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &pages[i]); err != nil {
			t.Fatal(err)
		}
		if pages[i].Total != 52 {
			t.Fatalf("incorrect total: %d", pages[i].Total)
		}
	}
	if len(pages[0].Data) != 50 || len(pages[1].Data) != 2 {
		t.Fatal("incorrect pagination")
	}
	if pages[0].Data[0].OwnerID != fmt.Sprintf("%024x", 51) || pages[1].Data[1].OwnerID != fmt.Sprintf("%024x", 52) {
		t.Fatal("distance must precede recency, including beyond the first reader page")
	}
	owners = []map[string]any{}
	rec := call(1, true)
	if rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	var empty struct {
		Total int
		Data  []lending.Copy
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &empty); err != nil || empty.Total != 0 || len(empty.Data) != 0 {
		t.Fatal("empty eligibility leaked copies")
	}
	for _, code := range []int{400, 401, 500} {
		status = code
		expected := code
		if code == 500 {
			expected = 503
		}
		if rec := call(1, true); rec.Code != expected {
			t.Fatalf("upstream %d: got %d", code, rec.Code)
		}
	}
}
