//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/auth"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/lending"
)

const bddSecret = "bdd-test-secret"

func newLendingRouter(s *lending.Service) http.Handler {
	lending.New(s)
	r := gin.New()
	s.Register(r.Group("/api"), r.Group("/api", auth.Required(bddSecret)))
	return r
}

func (s *lendingScenario) httpCall(actor, method, path string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(method, path, nil)
	if actor != "" {
		id, err := reader(actor)
		if err != nil {
			return nil, err
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{Email: "bdd@example.com", RegisteredClaims: jwt.RegisteredClaims{
			Subject: id, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}}).SignedString([]byte(bddSecret))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec, nil
}

func (s *lendingScenario) availability(expected string) error {
	rec, err := s.httpCall("Alice", "GET", "/api/me/copies")
	if err != nil {
		return err
	}
	if rec.Code != http.StatusOK {
		return fmt.Errorf("shelf returned %d: %s", rec.Code, rec.Body)
	}
	var body struct {
		Data []lending.Copy `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		return err
	}
	for _, copy := range body.Data {
		if copy.ID == s.copy.ID {
			if copy.Availability != expected {
				return fmt.Errorf("expected availability %q, got %q", expected, copy.Availability)
			}
			return nil
		}
	}
	return fmt.Errorf("copy missing from owner's shelf")
}

func (s *lendingScenario) archive(actor string) error {
	if err := s.snapshot(); err != nil {
		return err
	}
	rec, err := s.httpCall(actor, "DELETE", "/api/copies/"+s.copy.ID.String())
	if err != nil {
		return err
	}
	s.lastErr = nil
	switch rec.Code {
	case http.StatusNoContent:
	case http.StatusConflict:
		s.lastErr = lending.ErrConflict
	case http.StatusNotFound:
		s.lastErr = lending.ErrForbidden
	default:
		return fmt.Errorf("archive returned %d: %s", rec.Code, rec.Body)
	}
	return nil
}

func (s *lendingScenario) archiveRetainsHistory() error {
	var copy lending.Copy
	if err := testDB.First(&copy, "id = ?", s.copy.ID).Error; err != nil {
		return err
	}
	if !copy.Archived {
		return fmt.Errorf("copy was not archived")
	}
	for _, request := range s.requests {
		var persisted lending.Request
		if err := testDB.First(&persisted, "id = ?", request.ID).Error; err != nil {
			return err
		}
		var conversation lending.Conversation
		if err := testDB.First(&conversation, "request_id = ?", request.ID).Error; err != nil {
			return err
		}
		var message lending.Message
		if err := testDB.First(&message, "conversation_id = ?", conversation.ID).Error; err != nil {
			return err
		}
		if message.Body != "May I borrow this copy?" || message.SenderID != request.BorrowerID {
			return fmt.Errorf("archiving changed the introduction message")
		}
	}
	return nil
}

func (s *lendingScenario) absentPublicly() error {
	rec, err := s.httpCall("", "GET", "/api/copies")
	if err != nil {
		return err
	}
	if rec.Code != http.StatusOK {
		return fmt.Errorf("public shelf returned %d", rec.Code)
	}
	var body struct {
		Data []lending.Copy `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		return err
	}
	for _, copy := range body.Data {
		if copy.ID == s.copy.ID {
			return fmt.Errorf("archived copy exposed publicly")
		}
	}
	return nil
}

func (s *lendingScenario) registerHTTP(sc *godog.ScenarioContext) {
	sc.Step(`^the copy is "([^"]*)"$`, s.availability)
	sc.Step(`^(Alice|Bob|Carol) archives the copy$`, s.archive)
	sc.Step(`^the copy is archived with borrowing history retained$`, s.archiveRetainsHistory)
	sc.Step(`^the copy is absent from public shelves$`, s.absentPublicly)
}
