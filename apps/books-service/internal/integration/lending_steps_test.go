//go:build integration

package integration

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/lending"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
)

var bddReaders = map[string]string{
	"Alice": "0123456789abcdef01234567",
	"Bob":   "1123456789abcdef01234567",
	"Carol": "2123456789abcdef01234567",
}

type lendingScenario struct {
	service        *lending.Service
	router         http.Handler
	copy           lending.Copy
	requests       map[string]lending.Request
	lastErr        error
	beforeRequests []lending.Request
	beforeCopy     lending.Copy
}

func (s *lendingScenario) reset() error {
	*s = lendingScenario{requests: make(map[string]lending.Request)}
	if err := truncateTestDB(); err != nil {
		return err
	}
	s.service = &lending.Service{DB: testDB}
	s.router = newLendingRouter(s.service)
	return nil
}

func reader(name string) (string, error) {
	id, ok := bddReaders[name]
	if !ok {
		return "", fmt.Errorf("unknown reader %q", name)
	}
	return id, nil
}

func (s *lendingScenario) availableCopy() error {
	author := model.Author{Name: "BDD writer"}
	if err := testDB.Create(&author).Error; err != nil {
		return err
	}
	book := model.Book{Title: "BDD book", AuthorID: author.ID}
	if err := testDB.Create(&book).Error; err != nil {
		return err
	}
	s.copy = lending.Copy{ID: uuid.New(), BookID: book.ID, OwnerID: bddReaders["Alice"], Condition: "Good", Visible: true, Lendable: true}
	return testDB.Create(&s.copy).Error
}

func (s *lendingScenario) request(name string) error {
	id, err := reader(name)
	if err != nil {
		return err
	}
	r, err := s.service.RequestCopy(id, s.copy.ID, "May I borrow this copy?")
	if err == nil {
		s.requests[name] = r
	}
	return err
}

func (s *lendingScenario) snapshot() error {
	if err := testDB.Order("id").Find(&s.beforeRequests).Error; err != nil {
		return err
	}
	return testDB.First(&s.beforeCopy, "id = ?", s.copy.ID).Error
}

func (s *lendingScenario) attemptRequest(name string) error {
	if err := s.snapshot(); err != nil {
		return err
	}
	s.lastErr = s.request(name)
	return nil
}

var lendingActions = map[string]string{
	"accepts": "accept", "declines": "decline", "cancels": "cancel",
	"confirms handover of": "handover", "initiates return of": "return",
	"confirms receipt of": "confirm-return",
}

func (s *lendingScenario) transition(actor, action, borrower string) error {
	id, err := reader(actor)
	if err != nil {
		return err
	}
	r, ok := s.requests[borrower]
	if !ok {
		return fmt.Errorf("no request for %s", borrower)
	}
	if err := s.snapshot(); err != nil {
		return err
	}
	s.lastErr = s.service.Transition(id, r.ID, lendingActions[action], nil)
	return nil
}

func (s *lendingScenario) loanInState(status string) error {
	if err := s.request("Bob"); err != nil {
		return err
	}
	if status == "pending" {
		return nil
	}
	steps := []struct{ actor, action, target string }{
		{"Alice", "accepts", "accepted"}, {"Alice", "confirms handover of", "borrowed"},
		{"Bob", "initiates return of", "return_pending"}, {"Alice", "confirms receipt of", "returned"},
	}
	for _, step := range steps {
		if err := s.transition(step.actor, step.action, "Bob"); err != nil {
			return err
		}
		if s.lastErr != nil {
			return s.lastErr
		}
		if status == step.target {
			return nil
		}
	}
	return fmt.Errorf("unsupported setup status %q", status)
}

func (s *lendingScenario) succeeds() error {
	if s.lastErr != nil {
		return fmt.Errorf("expected success: %w", s.lastErr)
	}
	return nil
}

func (s *lendingScenario) rejected() error {
	if !errors.Is(s.lastErr, lending.ErrConflict) && !errors.Is(s.lastErr, lending.ErrForbidden) {
		return fmt.Errorf("expected a domain rejection, got %v", s.lastErr)
	}
	return s.unchanged()
}

func (s *lendingScenario) succeedsUnchanged() error {
	if err := s.succeeds(); err != nil {
		return err
	}
	return s.unchanged()
}

func (s *lendingScenario) unchanged() error {
	var requests []lending.Request
	if err := testDB.Order("id").Find(&requests).Error; err != nil {
		return err
	}
	var copy lending.Copy
	if err := testDB.First(&copy, "id = ?", s.copy.ID).Error; err != nil {
		return err
	}
	if !reflect.DeepEqual(requests, s.beforeRequests) || !reflect.DeepEqual(copy, s.beforeCopy) {
		return fmt.Errorf("action changed persisted requests or copy")
	}
	return nil
}

func (s *lendingScenario) requestStatus(name, expected string) error {
	r, ok := s.requests[name]
	if !ok {
		return fmt.Errorf("no request for %s", name)
	}
	var persisted lending.Request
	if err := testDB.First(&persisted, "id = ?", r.ID).Error; err != nil {
		return err
	}
	if persisted.Status != expected {
		return fmt.Errorf("%s request: expected %s, got %s", name, expected, persisted.Status)
	}
	return nil
}

func (s *lendingScenario) register(sc *godog.ScenarioContext) {
	sc.Step(`^Alice owns an available lendable physical copy$`, s.availableCopy)
	sc.Step(`^(Bob|Carol) has a pending request for the copy$`, s.request)
	sc.Step(`^Bob's loan is "([^"]*)"$`, s.loanInState)
	sc.Step(`^(Alice|Bob|Carol) requests the copy$`, s.attemptRequest)
	sc.Step(`^(Alice|Bob|Carol) (accepts|declines|cancels|confirms handover of|initiates return of|confirms receipt of) (Bob|Carol)'s request$`, s.transition)
	sc.Step(`^the action succeeds$`, s.succeeds)
	sc.Step(`^the action succeeds without changing the copy or requests$`, s.succeedsUnchanged)
	sc.Step(`^the action is rejected without changing the copy or requests$`, s.rejected)
	sc.Step(`^(Bob|Carol)'s request is "([^"]*)"$`, s.requestStatus)
	s.registerHTTP(sc)
}
