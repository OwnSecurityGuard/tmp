package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	ID   string
	conn *websocket.Conn
	send chan []byte
}

func newWSConn(id string, conn *websocket.Conn) *WSConn {
	return &WSConn{
		ID:   id,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (c *WSConn) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(1024 * 1024)

	//c.conn.SetReadDeadline(time.Now().Add(h.pingInterval * 2))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(h.pingInterval * 2))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			log.Printf("ws read error [%s]: %v", c.ID, err)
			return
		}
	}
}

func (c *WSConn) writePump(h *Hub) {
	ticker := time.NewTicker(h.pingInterval)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("ws write error [%s]: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ws ping failed [%s]: %v", c.ID, err)
				return
			}
		}
	}
}
