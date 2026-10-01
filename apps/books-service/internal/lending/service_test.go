package lending

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/auth"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const owner = "0123456789abcdef01234567"
const borrower = "1123456789abcdef01234567"
const stranger = "2123456789abcdef01234567"
const secret = "lending-test-secret"

func fixture(t *testing.T) (*Service, Copy) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared&_foreign_keys=on"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	// Verify upgrade from an existing catalog, including repeated migration.
	if err := db.AutoMigrate(&model.Author{}, &model.Book{}); err != nil {
		t.Fatal(err)
	}
	a := model.Author{Name: "Writer"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	b := model.Book{Title: "A real book", AuthorID: a.ID}
	if err := db.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	c := Copy{ID: uuid.New(), BookID: b.ID, OwnerID: owner, Condition: "Good", Visible: true, Lendable: true}
	if err := db.Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	New(s)
	return s, c
}

func TestBorrowingLifecycleAndInvariants(t *testing.T) {
	s, c := fixture(t)
	if _, err := s.RequestCopy(owner, c.ID, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("self borrowing: %v", err)
	}
	r, err := s.RequestCopy(borrower, c.ID, "May I borrow this?")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RequestCopy(borrower, c.ID, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate request: %v", err)
	}
	competing, err := s.RequestCopy(stranger, c.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Transition(borrower, r.ID, "accept", nil); !errors.Is(err, ErrConflict) {
		t.Fatalf("borrower approved: %v", err)
	}
	if err := s.Transition(stranger, r.ID, "accept", nil); !errors.Is(err, ErrForbidden) {
		t.Fatalf("stranger approved: %v", err)
	}
	for _, step := range []struct{ user, action string }{{owner, "accept"}, {owner, "handover"}, {borrower, "return"}, {owner, "confirm-return"}} {
		if err := s.Transition(step.user, r.ID, step.action, nil); err != nil {
			t.Fatalf("%s: %v", step.action, err)
		}
		if err := s.Transition(step.user, r.ID, step.action, nil); err != nil {
			t.Fatalf("retry %s: %v", step.action, err)
		}
		s.hydrateCopy(&c)
		if step.action == "return" && c.Availability != "on loan" {
			t.Fatal("borrower return released copy before receipt")
		}
	}
	s.DB.First(&competing, "id = ?", competing.ID)
	if competing.Status != "declined" {
		t.Fatal("competing request was not declined")
	}
	s.hydrateCopy(&c)
	if c.Availability != "available" {
		t.Fatal(c.Availability)
	}
	if err := s.Transition(owner, r.ID, "handover", nil); !errors.Is(err, ErrConflict) {
		t.Fatalf("reopened completed loan: %v", err)
	}
	if _, err := s.RequestCopy(stranger, c.ID, ""); err != nil {
		t.Fatal("copy cannot be lent again", err)
	}
	if err := s.DB.Delete(&model.Book{}, "id = ?", c.BookID).Error; err == nil {
		t.Fatal("deleted catalog title with physical copies")
	}
}

func TestMessagesPermissionsRequestsAndRetries(t *testing.T) {
	s, _ := fixture(t)
	c, err := s.Direct(owner, borrower)
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := s.Direct(borrower, owner)
	if err != nil || reverse.ID != c.ID {
		t.Fatal("duplicate pair conversation", err)
	}
	mid := uuid.New()
	if _, err := s.Send(stranger, c.ID, mid, "Intrusion"); !errors.Is(err, ErrForbidden) {
		t.Fatal("stranger send", err)
	}
	if _, err := s.Send(borrower, c.ID, mid, "Reply"); !errors.Is(err, ErrConflict) {
		t.Fatal("recipient sent before acceptance", err)
	}
	m, err := s.Send(owner, c.ID, mid, "Hello")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.Send(owner, c.ID, mid, "Hello")
	if err != nil || retry.ID != m.ID {
		t.Fatal("retry failed", err)
	}
	if _, err := s.Send(owner, c.ID, mid, "Changed"); !errors.Is(err, ErrConflict) {
		t.Fatal("changed retry", err)
	}
	if _, err := s.Send(owner, c.ID, uuid.New(), "Second introduction"); !errors.Is(err, ErrConflict) {
		t.Fatal("multiple introductions", err)
	}
	s.DB.Model(&c).Update("status", "accepted")
	if _, err := s.Send(borrower, c.ID, uuid.New(), "Hi!"); err != nil {
		t.Fatal(err)
	}
	s.DB.Create(&Block{OwnerID: borrower, UserID: owner})
	if _, err := s.Send(owner, c.ID, uuid.New(), "Blocked"); !errors.Is(err, ErrConflict) {
		t.Fatal("blocked sender", err)
	}
	if _, err := s.Direct(owner, borrower); !errors.Is(err, ErrConflict) {
		t.Fatal("blocked direct creation", err)
	}
	var count int64
	s.DB.Model(&Message{}).Where("conversation_id = ?", c.ID).Count(&count)
	if count != 2 {
		t.Fatalf("persisted messages: %d", count)
	}
}

func router(s *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s.Register(r.Group("/api"), r.Group("/api", auth.Required(secret)))
	return r
}
func token(user string) string {
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{Email: "private@example.com", RegisteredClaims: jwt.RegisteredClaims{Subject: user, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString([]byte(secret))
	return t
}
func call(t *testing.T, r http.Handler, user, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	if user != "" {
		req.Header.Set("Authorization", "Bearer "+token(user))
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHTTPPrivacyAndShelfOwnership(t *testing.T) {
	s, c := fixture(t)
	r := router(s)
	for _, path := range []string{"/api/me/copies", "/api/requests", "/api/conversations", "/api/blocks"} {
		if got := call(t, r, "", "GET", path, nil).Code; got != 401 {
			t.Fatalf("anonymous %s: %d", path, got)
		}
	}
	conv, _ := s.Direct(owner, borrower)
	s.Send(owner, conv.ID, uuid.New(), "Private")
	if got := call(t, r, stranger, "GET", "/api/conversations/"+conv.ID.String()+"/messages", nil).Code; got != 404 {
		t.Fatalf("private history: %d", got)
	}
	if got := call(t, r, stranger, "DELETE", "/api/copies/"+c.ID.String(), nil).Code; got != 404 {
		t.Fatalf("nonowner archive: %d", got)
	}
	s.DB.Model(&c).Update("visible", false)
	public := call(t, r, "", "GET", "/api/copies", nil)
	if public.Code != 200 || strings.Contains(public.Body.String(), c.ID.String()) {
		t.Fatal("hidden copy exposed", public.Body.String())
	}
	mine := call(t, r, owner, "GET", "/api/me/copies", nil)
	if mine.Code != 200 || !strings.Contains(mine.Body.String(), c.ID.String()) {
		t.Fatal("owner cannot see copy", mine.Body.String())
	}
	s.DB.Model(&c).Update("visible", true)
	request, _ := s.RequestCopy(borrower, c.ID, "")
	s.Transition(owner, request.ID, "accept", nil)
	if got := call(t, r, owner, "DELETE", "/api/copies/"+c.ID.String(), nil).Code; got != 409 {
		t.Fatalf("archived reservation: %d", got)
	}
	if got := call(t, r, borrower, "POST", "/api/conversations/"+conv.ID.String()+"/accept", map[string]any{}).Code; got != 204 {
		t.Fatalf("recipient acceptance: %d", got)
	}
	if got := call(t, r, owner, "GET", "/api/conversations", nil).Code; got != 200 {
		t.Fatal(got)
	}
}

func TestWebSocketTicketsDeliveryAndUnauthorizedEvents(t *testing.T) {
	s, _ := fixture(t)
	r := router(s)
	server := httptest.NewServer(r)
	defer server.Close()
	conv, _ := s.Direct(owner, borrower)
	s.DB.Model(&conv).Update("status", "accepted")
	origin := "http://localhost:5173"
	getTicket := func(user string) string {
		response := call(t, r, user, "POST", "/api/ws-ticket", map[string]string{"origin": origin})
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		var result map[string]string
		json.Unmarshal(response.Body.Bytes(), &result)
		return result["ticket"]
	}
	dial := func(user string) *websocket.Conn {
		ticket := getTicket(user)
		url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws?ticket=" + ticket
		conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Origin": []string{origin}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		if duplicate, response, err := websocket.DefaultDialer.Dial(url, http.Header{"Origin": []string{origin}}); err == nil {
			duplicate.Close()
			t.Fatal("ticket reused")
		} else if response.StatusCode != 401 {
			t.Fatal(response.StatusCode)
		}
		return conn
	}
	a, b := dial(owner), dial(borrower)
	mid := uuid.New()
	if err := a.WriteJSON(map[string]any{"event": "message", "conversationId": conv.ID, "id": mid, "body": "Instant exchange"}); err != nil {
		t.Fatal(err)
	}
	wait := func(conn *websocket.Conn, event string) map[string]any {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			var e struct {
				Event string         `json:"event"`
				Data  map[string]any `json:"data"`
			}
			if err := conn.ReadJSON(&e); err != nil {
				t.Fatal(err)
			}
			if e.Event == event {
				return e.Data
			}
		}
	}
	if data := wait(b, "message"); data["id"] != mid.String() {
		t.Fatal(data)
	}
	if data := wait(a, "ack"); data["id"] != mid.String() {
		t.Fatal(data)
	}
	outsider := dial(stranger)
	outsider.WriteJSON(map[string]any{"event": "message", "conversationId": conv.ID, "id": uuid.New(), "body": "Intrude"})
	wait(outsider, "error")
	ticket := getTicket(owner)
	if conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/api/ws?ticket="+ticket, http.Header{"Origin": []string{"https://wrong.example"}}); err == nil {
		conn.Close()
		t.Fatal("wrong origin connected")
	} else if response.StatusCode != 401 {
		t.Fatal(response.StatusCode)
	}
}
