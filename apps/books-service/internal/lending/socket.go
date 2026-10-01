package lending

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type socketTicket struct {
	user, origin            string
	expires, sessionExpires time.Time
}
type client struct {
	user string
	conn *websocket.Conn
	out  chan []byte
}
type Hub struct {
	mu      sync.Mutex
	tickets map[string]socketTicket
	clients map[string]map[*client]bool
	service *Service
}

func New(dbService *Service) *Hub {
	h := &Hub{tickets: map[string]socketTicket{}, clients: map[string]map[*client]bool{}, service: dbService}
	dbService.Hub = h
	return h
}

// @Summary Obtain a single-use WebSocket connection ticket
// @Description Connect to /api/ws?ticket=... within 30 seconds. The Origin must match the ticket. Events: message, requests, conversations, accept, decline, read, presence, typing, blocked, ack, error. Send message with conversationId, id and body; presence/typing with conversationId. Connections expire with the JWT.
// @Tags Messaging
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param origin body map[string]string true "Browser origin"
// @Success 200 {object} map[string]string
// @Router /ws-ticket [post]
func (h *Hub) ticket(c *gin.Context) {
	var in struct {
		Origin string `json:"origin"`
	}
	if !bind(c, &in) {
		return
	}
	u, err := url.Parse(in.Origin)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		c.JSON(400, gin.H{"message": "Invalid origin"})
		return
	}
	expires, ok := c.Get("expires_at")
	if !ok {
		c.Status(401)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, t := range h.tickets {
		if time.Now().After(t.expires) {
			delete(h.tickets, key)
		}
	}
	count := 0
	for _, t := range h.tickets {
		if t.user == c.GetString("user_id") {
			count++
		}
	}
	if count >= 5 {
		c.JSON(429, gin.H{"message": "Too many connection attempts"})
		return
	}
	key := uuid.NewString()
	h.tickets[key] = socketTicket{user: c.GetString("user_id"), origin: in.Origin, expires: time.Now().Add(30 * time.Second), sessionExpires: expires.(time.Time)}
	c.JSON(200, gin.H{"ticket": key})
}

func (h *Hub) broadcast(users []string, event string, data any) {
	encoded, _ := json.Marshal(gin.H{"event": event, "data": data})
	h.mu.Lock()
	defer h.mu.Unlock()
	seen := map[string]bool{}
	for _, user := range users {
		if seen[user] {
			continue
		}
		seen[user] = true
		for c := range h.clients[user] {
			select {
			case c.out <- encoded:
			default:
				c.conn.Close()
			}
		}
	}
}
func (h *Hub) online(user string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients[user]) > 0
}
func (h *Hub) presence(user string) {
	var cs []Conversation
	h.service.DB.Where("(user_a = ? OR user_b = ?) AND status = 'accepted'", user, user).Find(&cs)
	for _, c := range cs {
		if !h.service.blocked(h.service.DB, c.UserA, c.UserB) {
			h.broadcast([]string{other(c, user)}, "presence", gin.H{"conversationId": c.ID, "userId": user, "online": h.online(user)})
		}
	}
}

func (h *Hub) connect(c *gin.Context) {
	h.mu.Lock()
	t, ok := h.tickets[c.Query("ticket")]
	delete(h.tickets, c.Query("ticket"))
	if !ok || time.Now().After(t.expires) || time.Now().After(t.sessionExpires) || c.GetHeader("Origin") != t.origin || len(h.clients[t.user]) >= 4 {
		h.mu.Unlock()
		c.Status(401)
		return
	}
	h.mu.Unlock()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == t.origin }}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	cl := &client{user: t.user, conn: conn, out: make(chan []byte, 64)}
	h.mu.Lock()

	if len(h.clients[t.user]) >= 4 {
		h.mu.Unlock()
		conn.Close()
		return
	}
	if h.clients[t.user] == nil {
		h.clients[t.user] = map[*client]bool{}
	}
	h.clients[t.user][cl] = true
	h.mu.Unlock()
	h.presence(t.user)
	done := make(chan struct{})
	defer func() {
		close(done)
		conn.Close()
		h.mu.Lock()
		delete(h.clients[t.user], cl)
		if len(h.clients[t.user]) == 0 {
			delete(h.clients, t.user)
		}
		h.mu.Unlock()
		h.presence(t.user)
	}()
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		expiry := time.NewTimer(time.Until(t.sessionExpires))
		defer expiry.Stop()
		for {
			select {
			case <-done:
				return
			case <-expiry.C:
				conn.Close()
				return
			case data := <-cl.out:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if conn.WriteMessage(websocket.TextMessage, data) != nil {
					conn.Close()
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if conn.WriteMessage(websocket.PingMessage, nil) != nil {
					conn.Close()
					return
				}
			}
		}
	}()
	conn.SetReadLimit(8192)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	window, count := time.Now(), 0
	for {
		var in struct {
			Event          string    `json:"event"`
			ConversationID uuid.UUID `json:"conversationId"`
			ID             uuid.UUID `json:"id"`
			Body           string    `json:"body"`
		}
		if conn.ReadJSON(&in) != nil {
			return
		}
		if time.Since(window) > time.Second {
			window, count = time.Now(), 0
		}
		count++
		if count > 20 {
			return
		}
		if in.Event == "message" {
			m, err := h.service.Send(t.user, in.ConversationID, in.ID, in.Body)
			event, data := "ack", any(m)
			if err != nil {
				event, data = "error", gin.H{"id": in.ID, "message": "Message could not be sent. Check the conversation's permissions and status."}
			}
			encoded, _ := json.Marshal(gin.H{"event": event, "data": data})
			select {
			case cl.out <- encoded:
			default:
				return
			}
			continue
		}
		if in.Event == "typing" || in.Event == "presence" {
			var conversation Conversation
			if h.service.DB.First(&conversation, "id = ?", in.ConversationID).Error != nil || !participant(conversation, t.user) || conversation.Status != "accepted" || h.service.blocked(h.service.DB, conversation.UserA, conversation.UserB) {
				continue
			}
			if in.Event == "presence" {
				h.broadcast([]string{t.user}, "presence", gin.H{"conversationId": conversation.ID, "userId": other(conversation, t.user), "online": h.online(other(conversation, t.user))})
			} else {
				h.broadcast([]string{other(conversation, t.user)}, "typing", gin.H{"conversationId": conversation.ID, "userId": t.user})
			}
		}
	}
}
