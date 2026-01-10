package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	clients      map[string]*client
	clientByID   map[string]*client
	tokenMap     map[string]string // token -> clientID 映射
	register     chan *client
	unregister   chan *client
	mu           sync.RWMutex
	pingInterval time.Duration
	config       WebSocketConfig
}

// client 内部客户端结构
type client struct {
	*Client
	conn *websocket.Conn
	send chan []byte // 写队列（唯一写入口）
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService(config WebSocketConfig) *WebSocketService {
	return &WebSocketService{
		clients:      make(map[string]*client),
		clientByID:   make(map[string]*client),
		tokenMap:     make(map[string]string),
		register:     make(chan *client),
		unregister:   make(chan *client),
		pingInterval: config.PingInterval,
		config:       config,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// generateClientID 生成唯一客户端ID
func generateClientID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "client-" + time.Now().Format("20060102150405")
	}
	return "client-" + hex.EncodeToString(bytes)
}

// generateToken 生成客户端访问令牌
func generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "token-" + time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(bytes)
}

// HandleWebSocket 处理WebSocket连接升级
func (ws *WebSocketService) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("❌ WebSocket upgrade error:", err)
		return
	}

	clientID := generateClientID()
	token := generateToken()
	client := &client{
		Client: &Client{
			ID:    clientID,
			Token: token,
		},
		conn: conn,
		send: make(chan []byte, 256),
	}

	// 将token映射添加到tokenMap
	ws.mu.Lock()
	ws.tokenMap[token] = clientID
	ws.mu.Unlock()

	ws.register <- client

	go ws.readPump(client)
	go ws.writePump(client)

	// 发送欢迎消息给客户端，包含client_id和token
	welcomeMsg := map[string]interface{}{
		"type":      "welcome",
		"client_id": clientID,
		"token":     token,
		"message":   "连接成功",
		"timestamp": time.Now().Unix(),
	}

	jsonData, _ := json.Marshal(welcomeMsg)
	select {
	case client.send <- jsonData:
	default:
		log.Printf("Failed to send welcome message to client: %s", clientID)
	}

	log.Printf("✅ New client connected with ID: %s", clientID)
}

// readPump 处理客户端读取消息
func (ws *WebSocketService) readPump(client *client) {
	defer func() {
		ws.unregister <- client
		client.conn.Close()
	}()

	conn := client.conn
	conn.SetReadLimit(1024 * 1024)
	conn.SetReadDeadline(time.Now().Add(ws.pingInterval * 2))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(ws.pingInterval * 2))
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("read error, client disconnected: %v", err)
			return
		}
	}
}

// writePump 处理客户端写入消息
func (ws *WebSocketService) writePump(client *client) {
	ticker := time.NewTicker(ws.pingInterval)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	conn := client.conn

	for {
		select {
		case message, ok := <-client.send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("write error: %v", err)
				return
			}

		case <-ticker.C:
			//conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ping failed: %v", err)
				return
			}
		}
	}
}

// Run 启动WebSocket服务主循环
func (ws *WebSocketService) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.mu.Lock()
			ws.clients[client.ID] = client
			ws.clientByID[client.ID] = client
			ws.mu.Unlock()
			log.Printf("✅ client connected, ID: %s, current connections: %d", client.ID, len(ws.clients))

		case client := <-ws.unregister:
			ws.mu.Lock()
			if _, ok := ws.clients[client.ID]; ok {
				delete(ws.clients, client.ID)
				delete(ws.clientByID, client.ID)
				// 清理token映射
				if client.Token != "" {
					delete(ws.tokenMap, client.Token)
				}
				close(client.send)
			}
			ws.mu.Unlock()
			log.Printf("❌ client disconnected, ID: %s, current connections: %d", client.ID, len(ws.clients))
		}
	}
}

// SendToClient 向指定客户端发送消息
func (ws *WebSocketService) SendToClient(clientID string, message []byte) bool {
	ws.mu.RLock()
	client, exists := ws.clientByID[clientID]
	ws.mu.RUnlock()

	if !exists {
		log.Printf("client not found: %s", clientID)
		return false
	}

	select {
	case client.send <- message:
		return true
	default:
		log.Printf("write queue full for client: %s", clientID)
		ws.mu.Lock()
		ws.unregister <- client
		ws.mu.Unlock()
		return false
	}
}

// GetClientCount 获取当前连接客户端数量
func (ws *WebSocketService) GetClientCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.clients)
}

// GetClientIDs 获取所有连接客户端的ID列表
func (ws *WebSocketService) GetClientIDs() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	clientIDs := make([]string, 0, len(ws.clients))
	for clientID := range ws.clients {
		clientIDs = append(clientIDs, clientID)
	}
	return clientIDs
}

// GetFirstClientID 获取第一个连接的客户端ID，如果没有连接则返回空字符串
func (ws *WebSocketService) GetFirstClientID() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for clientID := range ws.clients {
		return clientID
	}
	return ""
}

// GetClientIDByToken 通过token获取客户端ID
func (ws *WebSocketService) GetClientIDByToken(token string) (string, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	clientID, exists := ws.tokenMap[token]
	return clientID, exists
}

// GetClientInfoByToken 通过token获取客户端详细信息
func (ws *WebSocketService) GetClientInfoByToken(token string) (ClientInfo, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	clientID, exists := ws.tokenMap[token]
	if !exists {
		return ClientInfo{}, false
	}

	client, exists := ws.clientByID[clientID]
	if !exists {
		return ClientInfo{}, false
	}

	info := ClientInfo{
		ID:          client.ID,
		Token:       client.Token,
		Status:      "connected",
		ConnectedAt: time.Now(),
	}

	return info, true
}

// SendMessage 发送JSON消息给指定客户端
func (ws *WebSocketService) SendMessage(eventType string, data interface{}) {
	clientID := ws.GetFirstClientID()
	if clientID == "" {
		log.Println("No connected clients to send message")
		return
	}

	message := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	ws.SendToClient(clientID, jsonData)
}
