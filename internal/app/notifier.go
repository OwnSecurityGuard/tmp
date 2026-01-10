package app

import (
	"encoding/json"
	"log"
	"proxy-system-backend/internal/modules/websocket"
	"time"
)

type WSNotifier struct {
	hub *websocket.Hub
}

func NewWSNotifier(hub *websocket.Hub) *WSNotifier {
	return &WSNotifier{hub: hub}
}

func (n *WSNotifier) HandleEvent(e Event) {
	msg := map[string]any{
		"type":      e.Type,
		"data":      e.Data,
		"timestamp": time.Now().Unix(),
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws notifier marshal failed: %v", err)
		return
	}

	n.hub.Broadcast(b)
}
