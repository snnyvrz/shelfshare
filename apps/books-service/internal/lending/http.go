package lending

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) Register(public, private *gin.RouterGroup) {
	public.GET("/copies", s.copies)
	private.GET("/me/copies", s.myCopies)
	private.GET("/nearby/copies", s.nearbyCopies)
	private.POST("/copies", s.createCopy)
	private.PATCH("/copies/:id", s.editCopy)
	private.DELETE("/copies/:id", s.archiveCopy)
	private.POST("/copies/:id/requests", s.requestCopy)
	private.GET("/requests", s.requests)
	private.POST("/requests/:id/:action", s.transition)
	private.GET("/conversations", s.conversations)
	private.GET("/conversations/:id", s.conversation)
	private.GET("/request-conversations/:id", s.requestConversation)
	private.POST("/conversations", s.direct)
	private.GET("/conversations/:id/messages", s.messages)
	private.POST("/conversations/:id/messages", s.send)
	private.POST("/conversations/:id/:action", s.conversationAction)
	private.GET("/blocks", s.blocks)
	private.POST("/blocks/:user", s.block)
	private.DELETE("/blocks/:user", s.unblock)
	private.POST("/ws-ticket", s.Hub.ticket)
	public.GET("/ws", s.Hub.connect)
}

func fail(c *gin.Context, err error) {
	code, message := http.StatusInternalServerError, "The request could not be completed"
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, ErrForbidden):
		code, message = 404, "Resource not found"
	case errors.Is(err, ErrConflict):
		code, message = 409, err.Error()
	}
	c.JSON(code, gin.H{"message": message})
}
func id(c *gin.Context) (uuid.UUID, bool) {
	value, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid ID"})
		return uuid.Nil, false
	}
	return value, true
}
func bind(c *gin.Context, value any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	if c.ShouldBindJSON(value) != nil {
		c.JSON(400, gin.H{"message": "Invalid input"})
		return false
	}
	return true
}
func limit(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	if page > 100000 {
		page = 100000
	}
	return 50, (page - 1) * 50
}

// copies lists public physical copies or the current user's private shelf.
// @Summary List physical copies
// @Tags Shelves
// @Produce json
// @Param ownerId query string false "Shelf owner"
// @Param bookId query string false "Catalog book"
// @Param page query int false "Page (50 copies)"
// @Success 200 {object} map[string]interface{}
// @Router /copies [get]
func (s *Service) copies(c *gin.Context) {
	query := s.DB.Model(&Copy{}).Preload("Book.Author")
	if strings.Contains(c.FullPath(), "/me/") {
		query = query.Where("owner_id = ? AND archived = false", c.GetString("user_id"))
	} else {
		query = query.Where("visible = true AND archived = false")
	}
	if owner := c.Query("ownerId"); owner != "" {
		query = query.Where("owner_id = ?", owner)
	}
	if book := c.Query("bookId"); book != "" {
		b, err := uuid.Parse(book)
		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid book ID"})
			return
		}
		query = query.Where("book_id = ?", b)
	}
	s.listCopies(c, query.Order("created_at DESC").Order("id ASC"))
}

func (s *Service) listCopies(c *gin.Context, query *gorm.DB) {
	size, offset := limit(c)
	var total int64
	if err := query.Session(&gorm.Session{}).Select("copies.id").Count(&total).Error; err != nil {
		fail(c, err)
		return
	}
	copies := []Copy{}
	if err := query.Limit(size).Offset(offset).Find(&copies).Error; err != nil {
		fail(c, err)
		return
	}
	for i := range copies {
		s.hydrateCopy(&copies[i])
	}
	c.JSON(200, gin.H{"data": copies, "total": total})
}

// @Summary List your private shelf, including hidden copies
// @Tags Shelves
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /me/copies [get]
func (s *Service) myCopies(c *gin.Context) { s.copies(c) }

type copyInput struct {
	BookID    uuid.UUID `json:"bookId"`
	Condition string    `json:"condition"`
	Notes     string    `json:"notes"`
	Visible   bool      `json:"visible"`
	Lendable  bool      `json:"lendable"`
}

func validCopy(in copyInput) bool {
	return len(in.Notes) <= 2000 && len(in.Condition) <= 120 && strings.TrimSpace(in.Condition) != ""
}

// @Summary Add a physical copy to your shelf
// @Tags Shelves
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param copy body copyInput true "Copy details"
// @Success 201 {object} Copy
// @Router /copies [post]
func (s *Service) createCopy(c *gin.Context) {
	var in copyInput
	if !bind(c, &in) {
		return
	}
	if !validCopy(in) || in.BookID == uuid.Nil {
		c.JSON(400, gin.H{"message": "Choose a book and describe its condition (max 120 characters); notes max 2000"})
		return
	}
	var n int64
	s.DB.Table("books").Where("id = ?", in.BookID).Count(&n)
	if n == 0 {
		c.JSON(404, gin.H{"message": "Book not found"})
		return
	}
	cp := Copy{ID: uuid.New(), BookID: in.BookID, OwnerID: c.GetString("user_id"), Condition: in.Condition, Notes: in.Notes, Visible: in.Visible, Lendable: in.Lendable}
	if err := s.DB.Create(&cp).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, cp)
}

// @Summary Edit your physical copy (ownership and catalog title are immutable)
// @Tags Shelves
// @Security BearerAuth
// @Accept json
// @Param id path string true "Copy ID"
// @Param details body copyInput true "Condition, notes, visibility and lending preference"
// @Success 204
// @Router /copies/{id} [patch]
func (s *Service) editCopy(c *gin.Context) { s.changeCopy(c, false) }

// @Summary Archive your unoccupied copy, preserving history
// @Tags Shelves
// @Security BearerAuth
// @Param id path string true "Copy ID"
// @Success 204
// @Failure 409 {object} map[string]string "Copy is reserved or on loan"
// @Router /copies/{id} [delete]
func (s *Service) archiveCopy(c *gin.Context) { s.changeCopy(c, true) }
func (s *Service) changeCopy(c *gin.Context, archive bool) {
	value, ok := id(c)
	if !ok {
		return
	}
	var in copyInput
	if !archive {
		if !bind(c, &in) {
			return
		}
		if !validCopy(in) {
			c.JSON(400, gin.H{"message": "Invalid condition or notes"})
			return
		}
	}
	var declined []Request
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var cp Copy
		if err := locked(tx).First(&cp, "id = ? AND owner_id = ? AND archived = false", value, c.GetString("user_id")).Error; err != nil {
			return err
		}
		if archive || !in.Visible || !in.Lendable {
			var n int64
			tx.Model(&Request{}).Where("copy_id = ? AND status IN ?", value, []string{"accepted", "borrowed", "return_pending"}).Count(&n)
			if n > 0 {
				return ErrConflict
			}
			if err := tx.Where("copy_id = ? AND status = 'pending'", value).Order("borrower_id").Find(&declined).Error; err != nil {
				return err
			}
			for _, r := range declined {
				if err := lockPair(tx, r.OwnerID, r.BorrowerID); err != nil {
					return err
				}
			}
			if err := tx.Model(&Request{}).Where("copy_id = ? AND status = 'pending'", value).Updates(map[string]any{"status": "declined", "closed_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		if archive {
			return tx.Model(&cp).Update("archived", true).Error
		}
		return tx.Model(&cp).Updates(map[string]any{"condition": in.Condition, "notes": in.Notes, "visible": in.Visible, "lendable": in.Lendable}).Error
	})
	if err != nil {
		fail(c, err)
		return
	}
	for _, r := range declined {
		s.notify(r.OwnerID, r.BorrowerID, "requests", nil)
	}
	c.Status(204)
}

// @Summary Request to borrow a physical copy
// @Tags Borrowing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Copy ID"
// @Param request body map[string]string true "Request message"
// @Success 201 {object} Request
// @Router /copies/{id}/requests [post]
func (s *Service) requestCopy(c *gin.Context) {
	value, ok := id(c)
	if !ok {
		return
	}
	var in struct {
		Message string `json:"message"`
	}
	if !bind(c, &in) {
		return
	}
	if len(in.Message) > 4000 {
		c.JSON(400, gin.H{"message": "Message is too long"})
		return
	}
	r, err := s.RequestCopy(c.GetString("user_id"), value, in.Message)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, r)
}

// @Summary List your borrowing and lending requests
// @Tags Borrowing
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /requests [get]
func (s *Service) requests(c *gin.Context) {
	user := c.GetString("user_id")
	size, offset := limit(c)
	rs := []Request{}
	query := s.DB.Where("owner_id = ? OR borrower_id = ?", user, user)
	var total int64
	query.Model(&Request{}).Count(&total)
	if err := query.Preload("Copy.Book.Author").Order("updated_at DESC").Limit(size).Offset(offset).Find(&rs).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"data": rs, "total": total})
}

// @Summary Change borrowing status (accept, decline, cancel, handover, return, confirm-return)
// @Tags Borrowing
// @Security BearerAuth
// @Accept json
// @Param id path string true "Request ID"
// @Param action path string true "Action"
// @Param terms body map[string]string true "Optional RFC3339 dueAt"
// @Success 204
// @Router /requests/{id}/{action} [post]
func (s *Service) transition(c *gin.Context) {
	value, ok := id(c)
	if !ok {
		return
	}
	var in struct {
		DueAt *time.Time `json:"dueAt"`
	}
	if !bind(c, &in) {
		return
	}
	if in.DueAt != nil && in.DueAt.Before(time.Now()) {
		c.JSON(400, gin.H{"message": "Due date must be in the future"})
		return
	}
	if err := s.Transition(c.GetString("user_id"), value, c.Param("action"), in.DueAt); err != nil {
		fail(c, err)
		return
	}
	c.Status(204)
}

// @Summary List your private conversations and unread counts
// @Tags Messaging
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /conversations [get]
func (s *Service) conversations(c *gin.Context) {
	user := c.GetString("user_id")
	size, offset := limit(c)
	cs := []Conversation{}
	query := s.DB.Model(&Conversation{}).Where("user_a = ? OR user_b = ?", user, user)
	var total int64
	query.Count(&total)
	if err := query.Preload("Request.Copy.Book.Author").Order("updated_at DESC").Limit(size).Offset(offset).Find(&cs).Error; err != nil {
		fail(c, err)
		return
	}
	for i := range cs {
		read := cs[i].AReadAt
		if cs[i].UserB == user {
			read = cs[i].BReadAt
		}
		s.DB.Model(&Message{}).Where("conversation_id = ? AND sender_id <> ? AND created_at > ?", cs[i].ID, user, read).Count(&cs[i].Unread)
		cs[i].PeerReadAt = cs[i].BReadAt
		if cs[i].UserB == user {
			cs[i].PeerReadAt = cs[i].AReadAt
		}
		if cs[i].Status != "accepted" || s.blocked(s.DB, cs[i].UserA, cs[i].UserB) {
			cs[i].PeerReadAt = time.Time{}
		}
		var m Message
		if s.DB.Where("conversation_id = ?", cs[i].ID).Order("created_at DESC, id DESC").First(&m).Error == nil {
			cs[i].LastMessage = &m
		}
	}
	c.JSON(200, gin.H{"data": cs, "total": total})
}

// @Summary Read a conversation as a participant
// @Tags Messaging
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} Conversation
// @Router /conversations/{id} [get]
func (s *Service) conversation(c *gin.Context) { s.conversationDetail(c, false) }

// @Summary Find a borrowing conversation as a participant
// @Tags Messaging
// @Security BearerAuth
// @Param id path string true "Borrow request ID"
// @Success 200 {object} Conversation
// @Router /request-conversations/{id} [get]
func (s *Service) requestConversation(c *gin.Context) { s.conversationDetail(c, true) }
func (s *Service) conversationDetail(c *gin.Context, request bool) {
	value, ok := id(c)
	if !ok {
		return
	}
	user := c.GetString("user_id")
	var conv Conversation
	field := "id"
	if request {
		field = "request_id"
	}
	if err := s.DB.Preload("Request.Copy.Book.Author").Where("user_a = ? OR user_b = ?", user, user).First(&conv, field+" = ?", value).Error; err != nil {
		fail(c, err)
		return
	}
	conv.PeerReadAt = conv.BReadAt
	if conv.UserB == user {
		conv.PeerReadAt = conv.AReadAt
	}
	if conv.Status != "accepted" || s.blocked(s.DB, conv.UserA, conv.UserB) {
		conv.PeerReadAt = time.Time{}
	}
	c.JSON(200, conv)
}

// @Summary Start or find a direct conversation
// @Tags Messaging
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param recipient body map[string]string true "userId"
// @Success 200 {object} Conversation
// @Router /conversations [post]
func (s *Service) direct(c *gin.Context) {
	var in struct {
		UserID string `json:"userId"`
	}
	if !bind(c, &in) {
		return
	}
	conv, err := s.Direct(c.GetString("user_id"), in.UserID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, conv)
}

// @Summary Read private message history (newest page first)
// @Tags Messaging
// @Security BearerAuth
// @Produce json
// @Param id path string true "Conversation ID"
// @Param before query string false "RFC3339 timestamp cursor"
// @Success 200 {object} map[string]interface{}
// @Router /conversations/{id}/messages [get]
func (s *Service) messages(c *gin.Context) {
	value, ok := id(c)
	if !ok {
		return
	}
	var conv Conversation
	if err := s.DB.First(&conv, "id = ? AND (user_a = ? OR user_b = ?)", value, c.GetString("user_id"), c.GetString("user_id")).Error; err != nil {
		fail(c, err)
		return
	}
	query := s.DB.Where("conversation_id = ?", value)
	if before := c.Query("before"); before != "" {
		cursor, err := time.Parse(time.RFC3339Nano, before)
		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid cursor"})
			return
		}
		if rawID := c.Query("beforeId"); rawID != "" {
			cursorID, err := uuid.Parse(rawID)
			if err != nil {
				c.JSON(400, gin.H{"message": "Invalid cursor ID"})
				return
			}
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)", cursor, cursor, cursorID)
		} else {
			query = query.Where("created_at < ?", cursor)
		}
	}
	ms := []Message{}
	if err := query.Order("created_at DESC, id DESC").Limit(50).Find(&ms).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"data": ms})
}

// @Summary Send a message with a client-generated UUID (idempotent retry)
// @Tags Messaging
// @Security BearerAuth
// @Accept json
// @Param id path string true "Conversation ID"
// @Param message body map[string]string true "id and body"
// @Success 201 {object} Message
// @Router /conversations/{id}/messages [post]
func (s *Service) send(c *gin.Context) {
	value, ok := id(c)
	if !ok {
		return
	}
	var in struct {
		ID   uuid.UUID `json:"id"`
		Body string    `json:"body"`
	}
	if !bind(c, &in) {
		return
	}
	m, err := s.Send(c.GetString("user_id"), value, in.ID, in.Body)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, m)
}

// @Summary Accept/decline a message request or mark a conversation read
// @Tags Messaging
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param action path string true "accept, decline or read"
// @Success 204
// @Router /conversations/{id}/{action} [post]
func (s *Service) conversationAction(c *gin.Context) {
	value, ok := id(c)
	if !ok {
		return
	}
	user := c.GetString("user_id")
	var conv Conversation
	if err := s.DB.First(&conv, "id = ?", value).Error; err != nil {
		fail(c, err)
		return
	}
	if !participant(conv, user) {
		fail(c, ErrForbidden)
		return
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockPair(tx, conv.UserA, conv.UserB); err != nil {
			return err
		}
		if err := locked(tx).First(&conv, "id = ?", value).Error; err != nil {
			return err
		}
		if !participant(conv, user) {
			return ErrForbidden
		}
		switch c.Param("action") {
		case "read":
			field := "a_read_at"
			if conv.UserB == user {
				field = "b_read_at"
			}
			return tx.Model(&conv).UpdateColumn(field, time.Now().UTC()).Error
		case "accept", "decline":
			if conv.RequestID != nil || conv.Status != "pending" || conv.Initiator == user || s.blocked(tx, conv.UserA, conv.UserB) {
				return ErrConflict
			}
			status := "accepted"
			if c.Param("action") == "decline" {
				status = "declined"
			}
			return tx.Model(&conv).Update("status", status).Error
		default:
			return ErrConflict
		}
	})
	if err != nil {
		fail(c, err)
		return
	}
	if c.Param("action") != "read" || conv.Status == "accepted" && !s.blocked(s.DB, conv.UserA, conv.UserB) {
		s.notify(conv.UserA, conv.UserB, c.Param("action"), gin.H{"conversationId": value, "userId": user, "at": time.Now().UTC()})
	} else {
		s.notify(user, user, "read", gin.H{"conversationId": value, "userId": user})
	}
	c.Status(204)
}

// @Summary List readers you have blocked
// @Tags Messaging
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /blocks [get]
func (s *Service) blocks(c *gin.Context) {
	bs := []Block{}
	if err := s.DB.Where("owner_id = ?", c.GetString("user_id")).Find(&bs).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"data": bs})
}

// @Summary Block direct messages and presence sharing from a reader
// @Tags Messaging
// @Security BearerAuth
// @Param user path string true "User ID"
// @Success 204
// @Router /blocks/{user} [post]
func (s *Service) block(c *gin.Context) {
	user, recipient := c.GetString("user_id"), c.Param("user")
	if user == recipient || !validUserID(recipient) {
		c.JSON(400, gin.H{"message": "Invalid user"})
		return
	}
	b := Block{OwnerID: user, UserID: recipient}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockPair(tx, user, recipient); err != nil {
			return err
		}
		return tx.Where(b).FirstOrCreate(&b).Error
	})
	if err != nil {
		fail(c, err)
		return
	}
	s.notify(user, recipient, "blocked", nil)
	c.Status(204)
}

// @Summary Unblock a reader
// @Tags Messaging
// @Security BearerAuth
// @Param user path string true "User ID"
// @Success 204
// @Router /blocks/{user} [delete]
func (s *Service) unblock(c *gin.Context) {
	user, recipient := c.GetString("user_id"), c.Param("user")
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockPair(tx, user, recipient); err != nil {
			return err
		}
		return tx.Delete(&Block{}, "owner_id = ? AND user_id = ?", user, recipient).Error
	})
	if err != nil {
		fail(c, err)
		return
	}
	s.notify(user, recipient, "blocked", nil)
	c.Status(204)
}
