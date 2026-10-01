package lending

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestBlockingPreservesActiveLoanCoordinationAndArchiveHistory(t *testing.T) {
	s, cp := fixture(t)
	r, err := s.RequestCopy(borrower, cp.ID, "An introduction")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Transition(owner, r.ID, "accept", nil); err != nil {
		t.Fatal(err)
	}
	var c Conversation
	if err := s.DB.First(&c, "request_id = ?", r.ID).Error; err != nil {
		t.Fatal(err)
	}
	s.DB.Create(&Block{OwnerID: owner, UserID: borrower})
	if _, err := s.Send(borrower, c.ID, uuid.New(), "Let's arrange the exchange"); err != nil {
		t.Fatal("blocking prevented loan coordination", err)
	}
	for _, step := range []struct{ user, action string }{{owner, "handover"}, {borrower, "return"}, {owner, "confirm-return"}} {
		if err := s.Transition(step.user, r.ID, step.action, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Send(borrower, c.ID, uuid.New(), "Closed loan"); !errors.Is(err, ErrConflict) {
		t.Fatal("closed thread writable", err)
	}
	http := router(s)
	if rec := call(t, http, owner, "DELETE", "/api/copies/"+cp.ID.String(), nil); rec.Code != 204 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	if rec := call(t, http, borrower, "GET", "/api/conversations/"+c.ID.String()+"/messages", nil); rec.Code != 200 {
		t.Fatal("history lost", rec.Code)
	}
	var n int64
	s.DB.Model(&Request{}).Where("id = ?", r.ID).Count(&n)
	if n != 1 {
		t.Fatal("archiving removed request")
	}
}

func TestArchiveClosesPendingThreadsAndMessagesRemainPrivate(t *testing.T) {
	s, cp := fixture(t)
	r, err := s.RequestCopy(borrower, cp.ID, "Can I borrow this?")
	if err != nil {
		t.Fatal(err)
	}
	var c Conversation
	s.DB.First(&c, "request_id = ?", r.ID)
	if rec := call(t, router(s), owner, "DELETE", "/api/copies/"+cp.ID.String(), nil); rec.Code != 204 {
		t.Fatal(rec.Body.String())
	}
	s.DB.First(&r, "id = ?", r.ID)
	if r.Status != "declined" || r.ClosedAt == nil {
		t.Fatal("pending request not closed", r.Status)
	}
	if _, err := s.Send(borrower, c.ID, uuid.New(), "New message"); !errors.Is(err, ErrConflict) {
		t.Fatal("archived pending thread writable", err)
	}
	if rec := call(t, router(s), stranger, "GET", "/api/conversations/"+c.ID.String(), nil); rec.Code != 404 {
		t.Fatal("conversation leaked", rec.Code)
	}
}
