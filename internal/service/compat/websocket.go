package compat

import (
	"proxy-system-backend/internal/service/core/websocket"
)

// WebSocketService 兼容WebSocketService - 直接使用websocket.WebSocketService
type WebSocketService = websocket.WebSocketService

// NewWebSocketService 创建兼容的WebSocket服务 - 直接使用websocket.NewWebSocketService
func NewWebSocketService() *WebSocketService {
	return websocket.NewWebSocketService(websocket.WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingInterval:    30 * 1000000000, // 30秒
	})
}
