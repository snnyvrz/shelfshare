//go:build integration

package integration

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/lending"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
)

func TestConcurrentApprovalReservesExactlyOneCopy(t *testing.T) {
	resetDB(t)
	a := model.Author{Name: "Lending writer"}
	if err := testDB.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	b := model.Book{Title: "Concurrent lending", AuthorID: a.ID}
	if err := testDB.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	owner := "0123456789abcdef01234567"
	cp := lending.Copy{ID: uuid.New(), BookID: b.ID, OwnerID: owner, Condition: "Good", Visible: true, Lendable: true}
	if err := testDB.Create(&cp).Error; err != nil {
		t.Fatal(err)
	}
	s := &lending.Service{DB: testDB}
	r1, err := s.RequestCopy("1123456789abcdef01234567", cp.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := s.RequestCopy("2123456789abcdef01234567", cp.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, r := range []lending.Request{r1, r2} {
		wg.Add(1)
		go func(r lending.Request) { defer wg.Done(); <-start; results <- s.Transition(owner, r.ID, "accept", nil) }(r)
	}
	close(start)
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, lending.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes %d, conflicts %d", successes, conflicts)
	}
	var occupied, declined int64
	testDB.Model(&lending.Request{}).Where("copy_id = ? AND status = 'accepted'", cp.ID).Count(&occupied)
	testDB.Model(&lending.Request{}).Where("copy_id = ? AND status = 'declined'", cp.ID).Count(&declined)
	if occupied != 1 || declined != 1 {
		t.Fatalf("occupied %d declined %d", occupied, declined)
	}
	// The database constraint is a second defense, independent of the service.
	invalid := lending.Request{ID: uuid.New(), CopyID: cp.ID, OwnerID: owner, BorrowerID: "3123456789abcdef01234567", Status: "accepted"}
	if err := testDB.Create(&invalid).Error; err == nil {
		t.Fatal("partial index allowed two reservations")
	}
	if err := lending.Migrate(testDB); err != nil {
		t.Fatal("repeat migration", err)
	}
	var copies int64
	testDB.Model(&lending.Copy{}).Count(&copies)
	if copies != 1 {
		t.Fatal("migration lost shelf data")
	}
}
