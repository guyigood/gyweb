# WebSocket 模块使用说明

## 功能特点

- 支持多客户端连接管理
- 消息广播和点对点通信
- 自动心跳检测和连接维护
- 灵活的消息处理回调
- 连接生命周期管理
- 线程安全的连接池
- 🔒 支持 WSS（WebSocket Secure）安全连接
- 🛡️ 自签名证书和自定义证书支持

## 基本使用

### 1. 创建 WebSocket Hub

```go
// 创建 Hub
wsHub := websocket.NewHub()
go wsHub.Run() // 启动 Hub
```

### 2. 配置 WebSocket

```go
// 使用默认配置
config := websocket.DefaultConfig()

// 使用 TLS 配置
config := websocket.DefaultTLSConfig()

// 或自定义配置
config := &websocket.WebSocketConfig{
    PingPeriod:     54 * time.Second,
    PongWait:       60 * time.Second,
    WriteWait:      10 * time.Second,
    MaxMessageSize: 1024,
    EnableTLS:      true,
}

// 设置回调函数
config.OnConnect = func(conn *websocket.Connection) {
    log.Printf("用户连接: %s", conn.GetClientID())
}

config.OnDisconnect = func(conn *websocket.Connection) {
    log.Printf("用户断开: %s", conn.GetClientID())
}

config.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
    log.Printf("收到消息: %v", msg)
}
```

### 3. 安全配置

```go
// 设置安全的来源验证
websocket.SetSecureCheckOrigin()

// 或自定义来源验证
websocket.SetCheckOrigin(func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    // 实现自定义验证逻辑
    return isValidOrigin(origin)
})
```

### 4. 处理 WebSocket 连接

```go
r.GET("/ws", func(c *gyarn.Context) {
    // 获取客户端ID（可以从认证中间件或查询参数获取）
    clientID := c.Query("user_id")
    if clientID == "" {
        clientID = generateUniqueID()
    }
    
    // 处理 WebSocket 连接
    websocket.HandleWebSocket(c, wsHub, config, clientID)
})
```

### 5. 发送消息

```go
// 广播消息给所有连接
message := []byte(`{"type":"broadcast","data":"Hello World"}`)
wsHub.Broadcast(message)

// 发送消息给指定客户端
wsHub.SendToClient("user123", message)

// 获取在线用户列表
users := wsHub.GetConnections()
```

## 消息格式

WebSocket 消息使用 JSON 格式：

```json
{
    "type": "消息类型",
    "data": "消息内容（任意类型）",
    "from": "发送者ID",
    "to": "接收者ID（可选）",
    "timestamp": 1640995200
}
```

### 常用消息类型

- `chat`: 聊天消息
- `private_message`: 私聊消息
- `user_joined`: 用户加入
- `user_left`: 用户离开
- `ping`: 心跳检测
- `pong`: 心跳响应

## 高级用法

### 1. 与认证系统集成

```go
// 在 WebSocket 处理中获取用户信息
r.GET("/ws", func(c *gyarn.Context) {
    // 从认证中间件获取用户ID
    userID, exists := c.Get("user_id")
    if !exists {
        c.Unauthorized("请先登录")
        return
    }
    
    clientID := fmt.Sprintf("user_%v", userID)
    websocket.HandleWebSocket(c, wsHub, config, clientID)
})
```

### 2. 房间系统

```go
// 实现房间系统
type RoomManager struct {
    rooms map[string]*websocket.Hub
    mutex sync.RWMutex
}

func (rm *RoomManager) JoinRoom(roomID string, conn *websocket.Connection) {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()
    
    if room, exists := rm.rooms[roomID]; exists {
        room.register <- conn
    } else {
        // 创建新房间
        room := websocket.NewHub()
        go room.Run()
        rm.rooms[roomID] = room
        room.register <- conn
    }
}
```

### 3. 消息持久化

```go
config.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
    // 保存消息到数据库
    if msg.Type == "chat" {
        saveMessageToDB(msg)
    }
    
    // 转发消息
    if data, err := json.Marshal(msg); err == nil {
        wsHub.Broadcast(data)
    }
}
```

## 客户端示例

### JavaScript 客户端

```javascript
// 自动检测协议
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = protocol + '//' + window.location.host + '/ws';
const ws = new WebSocket(wsUrl);

// 或手动指定WSS
const ws = new WebSocket('wss://localhost:8443/ws');

ws.onopen = function() {
    console.log('🔒 安全连接建立');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('🔒 收到安全消息:', message);
};

// 发送消息
function sendMessage() {
    const message = {
        type: 'chat',
        data: {
            content: '你好，世界！'
        }
    };
    ws.send(JSON.stringify(message));
}

// 发送私聊消息
function sendPrivateMessage(targetUserID, content) {
    const message = {
        type: 'private_message',
        to: targetUserID,
        data: {
            content: content
        }
    };
    ws.send(JSON.stringify(message));
}
```

## 配置说明

- `PingPeriod`: 发送 ping 消息的间隔
- `PongWait`: 等待 pong 响应的超时时间
- `WriteWait`: 写入消息的超时时间
- `MaxMessageSize`: 最大消息大小限制

## 性能优化建议

1. 合理设置消息大小限制
2. 使用消息队列处理高并发
3. 定期清理无效连接
4. 考虑使用 Redis 等外部存储共享状态
5. 在生产环境中限制来源检查

## 安全注意事项

1. 验证客户端来源（修改 `CheckOrigin` 函数）
2. 实现适当的认证和授权
3. 限制消息频率和大小
4. 防止恶意连接和消息
5. 使用 HTTPS/WSS 加密传输

## WSS（安全连接）支持

### 1. 启动 HTTPS 服务器

#### 方式1: 使用自签名证书（开发环境）

```go
func main() {
    r := engine.New()
    
    // 配置 WebSocket
    wsConfig := websocket.DefaultTLSConfig()
    websocket.SetSecureCheckOrigin()
    
    // 启动 HTTPS 服务器（自动生成自签名证书）
    if err := r.RunAutoTLS(":8443"); err != nil {
        log.Fatal(err)
    }
}
```

#### 方式2: 使用自定义证书（生产环境）

```go
func main() {
    r := engine.New()
    
    // 配置 WebSocket
    wsConfig := websocket.DefaultTLSConfig()
    websocket.SetSecureCheckOrigin()
    
    // 使用自定义证书
    tlsConfig := &engine.TLSConfig{
        CertFile: "cert.pem",
        KeyFile:  "key.pem",
    }
    
    if err := r.RunTLS(":8443", tlsConfig); err != nil {
        log.Fatal(err)
    }
}
```

### 2. 生成证书

使用框架提供的证书生成工具：

```bash
# 生成默认证书（localhost）
go run tools/cert_generator.go

# 生成指定域名的证书
go run tools/cert_generator.go -host="example.com"

# 生成指定有效期的证书
go run tools/cert_generator.go -duration=8760h  # 1年

# 生成CA证书
go run tools/cert_generator.go -ca=true

# 自定义输出文件名
go run tools/cert_generator.go -cert="mycert.pem" -key="mykey.pem"
```

### 3. 客户端连接

#### JavaScript 客户端

```javascript
// 自动检测协议
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = protocol + '//' + window.location.host + '/ws';
const ws = new WebSocket(wsUrl);

// 或手动指定WSS
const ws = new WebSocket('wss://localhost:8443/ws');

ws.onopen = function() {
    console.log('🔒 安全连接建立');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('🔒 收到安全消息:', message);
};
```

## 安全特性

### 1. 来源验证

```go
// 严格的来源验证（推荐生产环境使用）
websocket.SetCheckOrigin(func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    allowedOrigins := []string{
        "https://yourdomain.com",
        "https://www.yourdomain.com",
    }
    
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false
})
```

### 2. 认证集成

```go
// 与JWT认证集成
r.GET("/ws", authMiddleware, func(c *gyarn.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.Unauthorized("请先登录")
        return
    }
    
    clientID := fmt.Sprintf("user_%v", userID)
    websocket.HandleWebSocket(c, wsHub, config, clientID)
})
```

### 3. 消息加密

```go
// 在消息处理中添加加密
config.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
    // 解密消息
    if encryptedData, ok := msg.Data.(string); ok {
        decryptedData, err := decrypt(encryptedData)
        if err == nil {
            msg.Data = decryptedData
        }
    }
    
    // 处理解密后的消息
    handleMessage(conn, msg)
}
```

## 部署建议

### 1. 开发环境

```go
// 使用自签名证书
r.RunAutoTLS(":8443")
```

### 2. 生产环境

```go
// 使用有效的SSL证书
tlsConfig := &engine.TLSConfig{
    CertFile: "/path/to/cert.pem",
    KeyFile:  "/path/to/key.pem",
}
r.RunTLS(":443", tlsConfig)
```

### 3. 负载均衡配置

如果使用负载均衡器（如 Nginx），需要配置 WebSocket 代理：

```nginx
upstream websocket {
    server 127.0.0.1:8443;
}

server {
    listen 443 ssl;
    server_name yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location /ws {
        proxy_pass https://websocket;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 安全检查清单

- [ ] 使用有效的SSL证书（生产环境）
- [ ] 启用严格的来源验证
- [ ] 实现用户认证和授权
- [ ] 限制消息大小和频率
- [ ] 启用日志记录和监控
- [ ] 定期更新证书
- [ ] 使用强密码和密钥
- [ ] 配置防火墙规则

## 常见问题

### Q: 浏览器显示"不安全连接"警告？
A: 这是因为使用了自签名证书。在开发环境中，可以点击"高级"->"继续访问"。生产环境应使用有效的SSL证书。

### Q: WSS连接失败？
A: 检查：
1. 证书文件是否存在且有效
2. 端口是否被占用
3. 防火墙是否开放对应端口
4. 来源验证是否正确配置

### Q: 如何在Docker中使用？
A: 确保容器内的证书文件路径正确，并暴露HTTPS端口：
```dockerfile
EXPOSE 8443
VOLUME ["/certs"]
```

### Q: 如何集成Let's Encrypt？
A: 可以使用autocert包自动获取和更新证书：
```go
import "golang.org/x/crypto/acme/autocert"

m := &autocert.Manager{
    Cache:      autocert.DirCache("certs"),
    Prompt:     autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist("yourdomain.com"),
}

server := &http.Server{
    Addr:      ":443",
    Handler:   r,
    TLSConfig: m.TLSConfig(),
}
server.ListenAndServeTLS("", "")
``` 