package lending

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrConflict = errors.New("This action is not available in the current state")
var ErrForbidden = errors.New("This resource is not accessible")

type Service struct {
	DB  *gorm.DB
	Hub *Hub
}

func locked(db *gorm.DB) *gorm.DB {
	if db.Dialector.Name() == "postgres" {
		return db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return db
}

func participant(c Conversation, user string) bool { return c.UserA == user || c.UserB == user }

// Pair locks serialize blocking with conversation creation and delivery.
func lockPair(tx *gorm.DB, a, b string) error {
	if a > b {
		a, b = b, a
	}
	if tx.Dialector.Name() == "postgres" {
		return tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", "pair:"+a+":"+b).Error
	}
	return nil
}
func other(c Conversation, user string) string {
	if c.UserA == user {
		return c.UserB
	}
	return c.UserA
}

func (s *Service) blocked(db *gorm.DB, a, b string) bool {
	var n int64
	result := db.Model(&Block{}).Where("(owner_id = ? AND user_id = ?) OR (owner_id = ? AND user_id = ?)", a, b, b, a).Count(&n)
	return result.Error != nil || n > 0
}

func (s *Service) hydrateCopy(c *Copy) {
	c.Availability = "unavailable"
	if c.Archived || !c.Visible || !c.Lendable {
		return
	}
	var r Request
	if s.DB.Where("copy_id = ? AND status IN ?", c.ID, []string{"accepted", "borrowed", "return_pending"}).First(&r).Error == nil {
		c.Availability = "on loan"
		if r.Status == "accepted" {
			c.Availability = "reserved"
		}
	} else {
		c.Availability = "available"
	}
}

func (s *Service) RequestCopy(user string, id uuid.UUID, message string) (Request, error) {
	r := Request{ID: uuid.New(), CopyID: id, BorrowerID: user, Status: "pending", Message: strings.TrimSpace(message)}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var c Copy
		if err := locked(tx).First(&c, "id = ?", id).Error; err != nil {
			return err
		}
		if err := lockPair(tx, c.OwnerID, user); err != nil {
			return err
		}
		if c.OwnerID == user || !c.Visible || !c.Lendable || c.Archived || s.blocked(tx, c.OwnerID, user) {
			return ErrConflict
		}
		var n int64
		tx.Model(&Request{}).Where("copy_id = ? AND (status IN ? OR (borrower_id = ? AND status = 'pending'))", id, []string{"accepted", "borrowed", "return_pending"}, user).Count(&n)
		if n > 0 {
			return ErrConflict
		}
		r.OwnerID = c.OwnerID
		if err := tx.Create(&r).Error; err != nil {
			return err
		}
		conv := Conversation{ID: uuid.New(), Key: "request:" + r.ID.String(), UserA: c.OwnerID, UserB: user, RequestID: &r.ID, Status: "accepted", Initiator: user}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		if r.Message != "" {
			return tx.Create(&Message{ID: uuid.New(), ConversationID: conv.ID, SenderID: user, Body: r.Message}).Error
		}
		return nil
	})
	if err == nil {
		s.notify(r.OwnerID, r.BorrowerID, "requests", nil)
	}
	return r, err
}

func (s *Service) Transition(user string, id uuid.UUID, action string, due *time.Time) error {
	var initial Request
	if err := s.DB.First(&initial, "id = ?", id).Error; err != nil {
		return err
	}
	if user != initial.OwnerID && user != initial.BorrowerID {
		return ErrForbidden
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var c Copy
		if err := locked(tx).First(&c, "id = ?", initial.CopyID).Error; err != nil {
			return err
		}
		if err := lockPair(tx, initial.OwnerID, initial.BorrowerID); err != nil {
			return err
		}
		var r Request
		if err := locked(tx).First(&r, "id = ?", id).Error; err != nil {
			return err
		}
		owner := user == r.OwnerID
		// Repeat delivery of a completed action is harmless for the same actor.
		repeated := map[string]string{"accept": "accepted", "decline": "declined", "cancel": "cancelled", "handover": "borrowed", "return": "return_pending", "confirm-return": "returned"}
		authorized := action == "cancel" || owner && (action == "accept" || action == "decline" || action == "handover" || action == "confirm-return") || !owner && action == "return"
		if authorized && repeated[action] == r.Status {
			return nil
		}
		target := ""
		switch action {
		case "accept":
			if owner && r.Status == "pending" && c.Visible && c.Lendable && !c.Archived {
				target = "accepted"
			}
		case "decline":
			if owner && r.Status == "pending" {
				target = "declined"
			}
		case "cancel":
			if r.Status == "pending" && !owner || r.Status == "accepted" {
				target = "cancelled"
			}
		case "handover":
			if owner && r.Status == "accepted" {
				target = "borrowed"
			}
		case "return":
			if !owner && r.Status == "borrowed" {
				target = "return_pending"
			}
		case "confirm-return":
			if owner && r.Status == "return_pending" {
				target = "returned"
			}
		}
		if target == "" {
			return ErrConflict
		}
		if target == "accepted" {
			var n int64
			tx.Model(&Request{}).Where("copy_id = ? AND status IN ?", c.ID, []string{"accepted", "borrowed", "return_pending"}).Count(&n)
			if n > 0 || s.blocked(tx, r.OwnerID, r.BorrowerID) {
				return ErrConflict
			}
			if err := tx.Model(&Request{}).Where("copy_id = ? AND id <> ? AND status = 'pending'", c.ID, id).Updates(map[string]any{"status": "declined", "updated_at": time.Now().UTC(), "closed_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		changes := map[string]any{"status": target, "updated_at": time.Now().UTC()}
		now := time.Now().UTC()
		switch target {
		case "accepted":
			changes["accepted_at"] = now
		case "borrowed":
			changes["borrowed_at"] = now
		case "return_pending":
			changes["return_pending_at"] = now
		case "returned", "declined", "cancelled":
			changes["closed_at"] = now
		}
		if due != nil && (target == "accepted" || target == "borrowed") {
			changes["due_at"] = due
		}
		return tx.Model(&r).Updates(changes).Error
	})
	if err == nil {
		// Notify all competing borrowers as acceptance can decline their requests.
		var rs []Request
		s.DB.Where("copy_id = ?", initial.CopyID).Find(&rs)
		for _, r := range rs {
			s.notify(r.OwnerID, r.BorrowerID, "requests", nil)
		}
	}
	return err
}

func (s *Service) Direct(user, recipient string) (Conversation, error) {
	if user == recipient || !validUserID(recipient) {
		return Conversation{}, ErrConflict
	}
	a, b := user, recipient
	if a > b {
		a, b = b, a
	}
	key := "direct:" + a + ":" + b
	var c Conversation
	err := s.DB.Transaction(func(tx *gorm.DB) error {

		if err := lockPair(tx, a, b); err != nil {
			return err
		}
		if s.blocked(tx, a, b) {
			return ErrConflict
		}
		c = Conversation{ID: uuid.New(), Key: key, UserA: a, UserB: b, Status: "pending", Initiator: user}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoNothing: true}).Create(&c).Error; err != nil {
			return err
		}

		c = Conversation{}
		return tx.Where("key = ?", key).First(&c).Error
	})
	if err == nil {
		s.notify(a, b, "conversations", c)
	}
	return c, err
}

func validUserID(id string) bool {
	if len(id) != 24 {
		return false
	}
	for _, c := range id {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func (s *Service) Send(user string, convID, messageID uuid.UUID, body string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 4000 || messageID == uuid.Nil {
		return Message{}, ErrConflict
	}
	m := Message{ID: messageID, ConversationID: convID, SenderID: user, Body: body}

	var initial Conversation
	if err := s.DB.First(&initial, "id = ?", convID).Error; err != nil {
		return m, err
	}
	if !participant(initial, user) {
		return m, ErrForbidden
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockPair(tx, initial.UserA, initial.UserB); err != nil {
			return err
		}
		var c Conversation
		if err := locked(tx).First(&c, "id = ?", convID).Error; err != nil {
			return err
		}
		if !participant(c, user) {
			return ErrForbidden
		}
		var existing Message
		if err := tx.First(&existing, "id = ?", messageID).Error; err == nil {
			if existing.SenderID != user || existing.ConversationID != convID || existing.Body != body {
				return ErrConflict
			}
			m = existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if c.Status == "declined" {
			return ErrConflict
		}
		// Blocking prevents direct messaging; active loan coordination stays available.
		if c.RequestID == nil && s.blocked(tx, c.UserA, c.UserB) {
			return ErrConflict
		}
		if c.RequestID != nil {
			var r Request
			if err := tx.First(&r, "id = ?", c.RequestID).Error; err != nil {
				return err
			}
			if r.Status == "returned" || r.Status == "declined" || r.Status == "cancelled" {
				return ErrConflict
			}
		}
		if c.Status == "pending" {
			var n int64
			tx.Model(&Message{}).Where("conversation_id = ?", c.ID).Count(&n)
			if c.Initiator != user || n > 0 {
				return ErrConflict
			}
		}
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		return tx.Model(&c).Update("updated_at", time.Now().UTC()).Error
	})
	if err == nil {
		var c Conversation
		s.DB.First(&c, "id = ?", convID)
		s.notify(c.UserA, c.UserB, "message", m)
	}
	return m, err
}

func (s *Service) notify(a, b, event string, data any) {
	if s.Hub != nil {
		s.Hub.broadcast([]string{a, b}, event, data)
	}
}
