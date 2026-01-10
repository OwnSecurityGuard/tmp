package websocket

import (
	"log"
	"proxy-system-backend/internal/modules/shared"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Hub struct {
	conns        map[string]*WSConn
	register     chan *WSConn
	unregister   chan *WSConn
	mu           sync.RWMutex
	pingInterval time.Duration
}

func NewHub(cfg Config) *Hub {
	return &Hub{
		conns:        make(map[string]*WSConn),
		register:     make(chan *WSConn),
		unregister:   make(chan *WSConn),
		pingInterval: cfg.PingInterval,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.conns[c.ID] = c
			h.mu.Unlock()
			c.conn.WriteMessage(123, []byte("ASDAS"))
			log.Printf("✅ ws connected: %s, total=%d", c.ID, len(h.conns))

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.conns[c.ID]; ok {
				delete(h.conns, c.ID)
				close(c.send)
			}
			h.mu.Unlock()
			log.Printf("❌ ws disconnected: %s, total=%d", c.ID, len(h.conns))
		}
	}
}

func (h *Hub) Handle(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	wsConn := newWSConn(shared.GenerateConnID(), conn)
	h.register <- wsConn

	go wsConn.readPump(h)
	go wsConn.writePump(h)
}

func (h *Hub) Send(connID string, msg []byte) bool {
	h.mu.RLock()
	c, ok := h.conns[connID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case c.send <- msg:
		return true
	default:
		//h.unregister <- c
		return false
	}
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.conns {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func (h *Hub) ConnIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.conns))
	for id := range h.conns {
		ids = append(ids, id)
	}
	return ids
}
