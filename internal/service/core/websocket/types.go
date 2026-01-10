package websocket

import "time"

// Client WebSocket客户端
type Client struct {
	ID    string
	Token string      // 客户端访问令牌
}

// ClientInfo 客户端信息
type ClientInfo struct {
	ID       string    `json:"id"`
	Token    string    `json:"token"`
	Status   string    `json:"status"`
	ConnectedAt time.Time `json:"connected_at"`
}

// WebSocketConfig WebSocket服务配置
type WebSocketConfig struct {
	ReadBufferSize  int           `json:"read_buffer_size"`
	WriteBufferSize int           `json:"write_buffer_size"`
	PingInterval    time.Duration `json:"ping_interval"`
}