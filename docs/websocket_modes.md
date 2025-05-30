# WebSocket 双模式支持说明

## 概述

GyWeb 框架的 WebSocket 模块完全支持在 HTTP 和 HTTPS 服务器上运行，提供：

- **HTTP 模式**: 使用 `ws://` 协议
- **HTTPS 模式**: 使用 `wss://` 协议（WebSocket Secure）

## 工作原理

WebSocket 连接的建立过程：

1. **HTTP 握手**: 客户端发送 HTTP 升级请求
2. **协议升级**: 服务器将 HTTP 连接升级为 WebSocket
3. **数据传输**: 使用 WebSocket 协议进行双向通信

无论是 HTTP 还是 HTTPS 服务器，WebSocket 升级过程都是相同的，只是底层传输层的加密情况不同。

## 使用方式

### 1. HTTP 模式 (ws://)

```go
func main() {
    r := engine.New()
    
    // 创建 WebSocket Hub
    wsHub := websocket.NewHub()
    go wsHub.Run()
    
    // 配置 WebSocket
    wsConfig := websocket.DefaultConfig()
    
    // WebSocket 路由
    r.GET("/ws", func(c *gyarn.Context) {
        websocket.HandleWebSocket(c, wsHub, wsConfig, "user123")
    })
    
    // 启动 HTTP 服务器
    r.Run(":8080")  // ws://localhost:8080/ws
}
```

### 2. HTTPS 模式 (wss://)

```go
func main() {
    r := engine.New()
    
    // 创建 WebSocket Hub
    wsHub := websocket.NewHub()
    go wsHub.Run()
    
    // 配置 WebSocket（TLS模式）
    wsConfig := websocket.DefaultTLSConfig()
    websocket.SetSecureCheckOrigin()
    
    // WebSocket 路由
    r.GET("/ws", func(c *gyarn.Context) {
        websocket.HandleWebSocket(c, wsHub, wsConfig, "user123")
    })
    
    // 启动 HTTPS 服务器
    r.RunAutoTLS(":8443")  // wss://localhost:8443/ws
    
    // 或使用自定义证书
    // tlsConfig := &engine.TLSConfig{
    //     CertFile: "cert.pem",
    //     KeyFile:  "key.pem",
    // }
    // r.RunTLS(":8443", tlsConfig)
}
```

### 3. 双模式运行

```go
func main() {
    // 同时启动 HTTP 和 HTTPS 服务器
    go func() {
        r1 := engine.New()
        setupWebSocket(r1)
        r1.Run(":8080")  // HTTP/WS
    }()
    
    r2 := engine.New()
    setupWebSocket(r2)
    r2.RunAutoTLS(":8443")  // HTTPS/WSS
}
```

## 客户端连接

### JavaScript 自动检测协议

```javascript
// 自动检测当前页面协议
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = protocol + '//' + window.location.host + '/ws';
const ws = new WebSocket(wsUrl);

ws.onopen = function() {
    console.log(protocol === 'wss:' ? '🔒 安全连接建立' : '🌐 普通连接建立');
};
```

### 指定协议连接

```javascript
// HTTP 模式
const wsHttp = new WebSocket('ws://localhost:8080/ws');

// HTTPS 模式
const wsHttps = new WebSocket('wss://localhost:8443/ws');
```

## 配置差异

### HTTP 模式配置

```go
config := websocket.DefaultConfig()
config.OnConnect = func(conn *websocket.Connection) {
    log.Printf("HTTP 连接: %s", conn.GetClientID())
}
```

### HTTPS 模式配置

```go
config := websocket.DefaultTLSConfig()  // 自动设置 EnableTLS = true
websocket.SetSecureCheckOrigin()       // 启用安全来源验证

config.OnConnect = func(conn *websocket.Connection) {
    log.Printf("HTTPS 连接: %s", conn.GetClientID())
}
```

## 安全考虑

| 模式 | 加密 | 来源验证 | 适用场景 |
|------|------|----------|----------|
| HTTP/WS | ❌ | 宽松 | 开发环境、内网应用 |
| HTTPS/WSS | ✅ | 严格 | 生产环境、公网应用 |

### 安全建议

1. **生产环境必须使用 WSS**
2. **启用严格来源验证**
3. **使用有效的 SSL 证书**
4. **实现用户认证**

## 常见问题

### Q: 能否在同一个应用中同时支持 ws:// 和 wss://?
A: 可以！只需要在不同端口上启动 HTTP 和 HTTPS 服务器即可。

### Q: WebSocket 升级是否影响性能？
A: 不会。WebSocket 升级只在连接建立时发生一次，之后就是原生的 WebSocket 协议。

### Q: 如何在负载均衡环境中使用？
A: 确保负载均衡器支持 WebSocket 代理，并正确配置 `Upgrade` 和 `Connection` 头。

## 示例项目

- `examples/websocket_example.go` - 基础 HTTP/WS 示例
- `examples/websocket_tls_example.go` - HTTPS/WSS 示例  
- `examples/websocket_dual_mode.go` - 双模式示例

## 测试方法

```bash
# 启动双模式服务器
go run examples/websocket_dual_mode.go

# 测试 HTTP 模式
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
     http://localhost:8080/ws

# 测试 HTTPS 模式（需要忽略证书验证）
curl -k -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
     https://localhost:8443/ws
``` 