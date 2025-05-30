package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyigood/gyweb/core/gyarn"
)

// Upgrader WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该验证来源
		return true
	},
}

// SetCheckOrigin 设置来源验证函数
func SetCheckOrigin(checkOrigin func(r *http.Request) bool) {
	upgrader.CheckOrigin = checkOrigin
}

// SetSecureCheckOrigin 设置安全的来源验证（仅允许相同域名）
func SetSecureCheckOrigin() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		// 检查协议和主机是否匹配
		scheme := "ws"
		if r.TLS != nil {
			scheme = "wss"
		}
		expectedOrigin := scheme + "://" + r.Host
		return origin == expectedOrigin ||
			origin == "http://"+r.Host ||
			origin == "https://"+r.Host
	}
}

// Connection WebSocket 连接封装
type Connection struct {
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	clientID string
	userID   interface{} // 用户ID，可以是任意类型
}

// Hub WebSocket 连接管理器
type Hub struct {
	connections map[string]*Connection
	broadcast   chan []byte
	register    chan *Connection
	unregister  chan *Connection
	mutex       sync.RWMutex
}

// Message WebSocket 消息结构
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	From      string      `json:"from,omitempty"`
	To        string      `json:"to,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// MessageHandler 消息处理函数类型
type MessageHandler func(*Connection, *Message)

// ConnectionHandler 连接处理函数类型
type ConnectionHandler func(*Connection)

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	PingPeriod     time.Duration     // Ping 间隔
	PongWait       time.Duration     // Pong 等待时间
	WriteWait      time.Duration     // 写入等待时间
	MaxMessageSize int64             // 最大消息大小
	OnConnect      ConnectionHandler // 连接建立时的回调
	OnDisconnect   ConnectionHandler // 连接断开时的回调
	OnMessage      MessageHandler    // 收到消息时的回调
	EnableTLS      bool              // 是否启用 TLS
	TLSConfig      *TLSConfig        // TLS 配置
}

// TLSConfig TLS 配置
type TLSConfig struct {
	CertFile string // 证书文件路径
	KeyFile  string // 私钥文件路径
}

// DefaultConfig 默认配置
func DefaultConfig() *WebSocketConfig {
	return &WebSocketConfig{
		PingPeriod:     54 * time.Second,
		PongWait:       60 * time.Second,
		WriteWait:      10 * time.Second,
		MaxMessageSize: 1024,
		EnableTLS:      false,
	}
}

// DefaultTLSConfig 默认 TLS 配置
func DefaultTLSConfig() *WebSocketConfig {
	config := DefaultConfig()
	config.EnableTLS = true
	return config
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}
}

// Run 启动Hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.connections[conn.clientID] = conn
			h.mutex.Unlock()
			log.Printf("WebSocket 客户端连接: %s", conn.clientID)

		case conn := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.connections[conn.clientID]; ok {
				delete(h.connections, conn.clientID)
				close(conn.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket 客户端断开: %s", conn.clientID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, conn := range h.connections {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(h.connections, conn.clientID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast 广播消息给所有连接
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToClient 发送消息给指定客户端
func (h *Hub) SendToClient(clientID string, message []byte) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if conn, ok := h.connections[clientID]; ok {
		select {
		case conn.send <- message:
			return true
		default:
			return false
		}
	}
	return false
}

// GetConnections 获取所有连接的客户端ID
func (h *Hub) GetConnections() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]string, 0, len(h.connections))
	for clientID := range h.connections {
		clients = append(clients, clientID)
	}
	return clients
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(c *gyarn.Context, hub *Hub, config *WebSocketConfig, clientID string) {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	// 创建连接对象
	connection := &Connection{
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      hub,
		clientID: clientID,
	}

	// 从上下文中获取用户信息
	if userID, exists := c.Get("user_id"); exists {
		connection.userID = userID
	}

	// 注册连接
	hub.register <- connection

	// 调用连接建立回调
	if config.OnConnect != nil {
		config.OnConnect(connection)
	}

	// 启动读写协程
	go connection.writePump(config)
	go connection.readPump(config)
}

// readPump 读取消息
func (c *Connection) readPump(config *WebSocketConfig) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		if config.OnDisconnect != nil {
			config.OnDisconnect(c)
		}
	}()

	c.conn.SetReadLimit(config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
		return nil
	})

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 错误: %v", err)
			}
			break
		}

		// 设置时间戳
		message.Timestamp = time.Now().Unix()
		message.From = c.clientID

		// 调用消息处理回调
		if config.OnMessage != nil {
			config.OnMessage(c, &message)
		}
	}
}

// writePump 发送消息
func (c *Connection) writePump(config *WebSocketConfig) {
	ticker := time.NewTicker(config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 发送队列中的其他消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 发送消息给客户端
func (c *Connection) SendMessage(message []byte) {
	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}

// GetClientID 获取客户端ID
func (c *Connection) GetClientID() string {
	return c.clientID
}

// GetUserID 获取用户ID
func (c *Connection) GetUserID() interface{} {
	return c.userID
}

// SetUserID 设置用户ID
func (c *Connection) SetUserID(userID interface{}) {
	c.userID = userID
}
