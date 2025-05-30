package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/websocket"
)

// 创建WebSocket处理逻辑
func createWebSocketConfig(wsHub *websocket.Hub) *websocket.WebSocketConfig {
	config := websocket.DefaultConfig()

	config.OnConnect = func(conn *websocket.Connection) {
		log.Printf("新用户连接: %s", conn.GetClientID())

		// 广播用户加入消息
		message := websocket.Message{
			Type: "user_joined",
			Data: map[string]interface{}{
				"user_id": conn.GetClientID(),
				"message": fmt.Sprintf("用户 %s 加入了聊天室", conn.GetClientID()),
			},
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
		}

		// 发送在线用户列表
		onlineUsers := wsHub.GetConnections()
		userListMessage := websocket.Message{
			Type: "online_users",
			Data: map[string]interface{}{
				"users": onlineUsers,
				"count": len(onlineUsers),
			},
			To:        conn.GetClientID(),
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(userListMessage); err == nil {
			conn.SendMessage(data)
		}
	}

	config.OnDisconnect = func(conn *websocket.Connection) {
		log.Printf("用户断开连接: %s", conn.GetClientID())

		// 广播用户离开消息
		message := websocket.Message{
			Type: "user_left",
			Data: map[string]interface{}{
				"user_id": conn.GetClientID(),
				"message": fmt.Sprintf("用户 %s 离开了聊天室", conn.GetClientID()),
			},
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
		}
	}

	config.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
		log.Printf("收到消息: %s -> %s: %v", msg.From, msg.Type, msg.Data)

		switch msg.Type {
		case "chat":
			// 广播聊天消息
			if data, err := json.Marshal(msg); err == nil {
				wsHub.Broadcast(data)
			}

		case "private_message":
			// 私聊消息
			if msg.To != "" {
				if data, err := json.Marshal(msg); err == nil {
					wsHub.SendToClient(msg.To, data)
				}
			}

		case "ping":
			// 心跳响应
			pongMessage := websocket.Message{
				Type:      "pong",
				Data:      "ok",
				To:        conn.GetClientID(),
				Timestamp: time.Now().Unix(),
			}
			if data, err := json.Marshal(pongMessage); err == nil {
				conn.SendMessage(data)
			}
		}
	}

	return config
}

// 创建路由和处理器
func setupRoutes(r *engine.Engine, wsHub *websocket.Hub, wsConfig *websocket.WebSocketConfig) {
	// WebSocket 端点
	r.GET("/ws", func(c *gyarn.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			userID = "user_" + strconv.FormatInt(time.Now().UnixNano(), 36)
		}

		websocket.HandleWebSocket(c, wsHub, wsConfig, userID)
	})

	// 主页 - 自动检测协议
	r.GET("/", func(c *gyarn.Context) {
		c.HTML(200, `
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket 双模式聊天室</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .protocol-badge { 
            padding: 2px 8px; 
            border-radius: 4px; 
            font-size: 12px; 
            margin-left: 10px; 
        }
        .http-badge { background: #17a2b8; color: white; }
        .https-badge { background: #28a745; color: white; }
        #messages { 
            height: 400px; 
            border: 1px solid #ddd; 
            overflow-y: scroll; 
            padding: 10px; 
            background: #f9f9f9; 
            border-radius: 4px;
        }
        .input-group { 
            margin: 10px 0; 
            display: flex; 
            gap: 10px; 
        }
        input[type="text"] { 
            flex: 1; 
            padding: 8px; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
        }
        button { 
            padding: 8px 16px; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
        }
        .btn-primary { background: #007bff; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .status { 
            padding: 10px; 
            margin: 10px 0; 
            border-radius: 4px; 
        }
        .status.connected { background: #d4edda; border: 1px solid #c3e6cb; }
        .status.disconnected { background: #f8d7da; border: 1px solid #f5c6cb; }
        .info-panel {
            background: #e9ecef;
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>WebSocket 双模式聊天室</h1>
        
        <div class="info-panel">
            <h3>连接信息</h3>
            <p><strong>当前协议:</strong> <span id="currentProtocol"></span></p>
            <p><strong>WebSocket URL:</strong> <span id="wsUrl"></span></p>
            <p><strong>访问方式:</strong></p>
            <ul>
                <li>HTTP: <a href="http://localhost:8080" target="_blank">http://localhost:8080</a> (使用 ws://)</li>
                <li>HTTPS: <a href="https://localhost:8443" target="_blank">https://localhost:8443</a> (使用 wss://)</li>
            </ul>
        </div>
        
        <div id="status" class="status">连接中...</div>
        
        <div id="messages"></div>
        
        <div class="input-group">
            <input type="text" id="messageInput" placeholder="输入消息...">
            <button class="btn-primary" onclick="sendMessage()">发送</button>
            <button class="btn-secondary" onclick="sendPrivateMessage()">私聊</button>
        </div>
        
        <div>在线用户: <span id="onlineUsers">0</span></div>
    </div>
    
    <script>
        // 自动检测协议
        const isSecure = window.location.protocol === 'https:';
        const wsProtocol = isSecure ? 'wss:' : 'ws:';
        const wsUrl = wsProtocol + '//' + window.location.host + '/ws?user_id=' + Math.random().toString(36).substr(2, 9);
        
        // 更新页面信息
        document.getElementById('currentProtocol').innerHTML = 
            isSecure ? 
            '<span class="protocol-badge https-badge">🔒 HTTPS/WSS</span>' : 
            '<span class="protocol-badge http-badge">🌐 HTTP/WS</span>';
        document.getElementById('wsUrl').textContent = wsUrl;
        
        const ws = new WebSocket(wsUrl);
        const messages = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const onlineUsers = document.getElementById('onlineUsers');
        const status = document.getElementById('status');
        
        ws.onopen = function() {
            const protocolText = isSecure ? '🔒 安全连接' : '🌐 普通连接';
            addMessage('系统', protocolText + '建立成功!', 'green');
            status.textContent = protocolText + ' (已连接)';
            status.className = 'status connected';
        };
        
        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            console.log('收到消息:', message);
            
            switch(message.type) {
                case 'chat':
                    const userBadge = isSecure ? '🔒' : '🌐';
                    addMessage(message.from + ' ' + userBadge, message.data.content, 'black');
                    break;
                case 'private_message':
                    const privateBadge = isSecure ? '🔒' : '🌐';
                    addMessage(message.from + ' ' + privateBadge + ' (私聊)', message.data.content, 'blue');
                    break;
                case 'user_joined':
                case 'user_left':
                    addMessage('系统', message.data.message, 'orange');
                    break;
                case 'online_users':
                    onlineUsers.textContent = message.data.count + ' 人在线';
                    break;
                case 'pong':
                    console.log('收到心跳响应');
                    break;
            }
        };
        
        ws.onclose = function() {
            const protocolText = isSecure ? '🔒 安全连接' : '🌐 普通连接';
            addMessage('系统', protocolText + '断开', 'red');
            status.textContent = protocolText + ' (断开)';
            status.className = 'status disconnected';
        };
        
        ws.onerror = function(error) {
            const protocolText = isSecure ? '🔒 安全连接' : '🌐 普通连接';
            addMessage('系统', protocolText + '错误', 'red');
            status.textContent = protocolText + ' (错误)';
            status.className = 'status disconnected';
        };
        
        function addMessage(from, content, color) {
            const div = document.createElement('div');
            div.style.color = color;
            div.style.marginBottom = '5px';
            div.innerHTML = '<strong>' + from + ':</strong> ' + content + 
                          ' <small style="color: #666;">(' + new Date().toLocaleTimeString() + ')</small>';
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }
        
        function sendMessage() {
            const content = messageInput.value.trim();
            if (content) {
                ws.send(JSON.stringify({
                    type: 'chat',
                    data: {
                        content: content,
                        protocol: isSecure ? 'wss' : 'ws'
                    }
                }));
                messageInput.value = '';
            }
        }
        
        function sendPrivateMessage() {
            const target = prompt('输入目标用户ID:');
            const content = messageInput.value.trim();
            if (target && content) {
                ws.send(JSON.stringify({
                    type: 'private_message',
                    to: target,
                    data: {
                        content: content,
                        protocol: isSecure ? 'wss' : 'ws'
                    }
                }));
                messageInput.value = '';
            }
        }
        
        messageInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
        
        // 定时发送心跳
        setInterval(function() {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    type: 'ping', 
                    data: {protocol: isSecure ? 'wss' : 'ws'}
                }));
            }
        }, 30000);
    </script>
</body>
</html>
		`)
	})

	// API 端点 - 获取在线用户
	r.GET("/api/online", func(c *gyarn.Context) {
		users := wsHub.GetConnections()
		c.Success(gyarn.H{
			"users": users,
			"count": len(users),
		})
	})

	// API 端点 - 发送广播消息
	r.POST("/api/broadcast", func(c *gyarn.Context) {
		var req struct {
			Message string `json:"message"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		message := websocket.Message{
			Type: "system_broadcast",
			Data: map[string]interface{}{
				"content": req.Message,
			},
			From:      "system",
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
			c.Success(gyarn.H{"message": "广播发送成功"})
		} else {
			c.InternalServerError("发送失败")
		}
	})
}

// 启动HTTP服务器
func startHTTPServer() {
	r := engine.New()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 创建 WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// WebSocket 配置
	wsConfig := createWebSocketConfig(wsHub)

	// 设置路由
	setupRoutes(r, wsHub, wsConfig)

	log.Println("🌐 HTTP 服务器启动: http://localhost:8080")
	log.Println("   - WebSocket: ws://localhost:8080/ws")

	if err := r.Run(":8080"); err != nil {
		log.Printf("HTTP 服务器启动失败: %v", err)
	}
}

// 启动HTTPS服务器
func startHTTPSServer() {
	r := engine.New()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 创建 WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// WebSocket 配置（TLS模式）
	wsConfig := createWebSocketConfig(wsHub)
	wsConfig.EnableTLS = true

	// 设置安全的来源验证
	websocket.SetSecureCheckOrigin()

	// 设置路由
	setupRoutes(r, wsHub, wsConfig)

	log.Println("🔒 HTTPS 服务器启动: https://localhost:8443")
	log.Println("   - WebSocket: wss://localhost:8443/ws")

	if err := r.RunAutoTLS(":8443"); err != nil {
		log.Printf("HTTPS 服务器启动失败: %v", err)
	}
}

// 双模式聊天室示例
func main() {
	log.Println("启动双模式 WebSocket 聊天室...")
	log.Println()

	// 同时启动 HTTP 和 HTTPS 服务器
	go startHTTPServer() // 在后台启动HTTP服务器
	startHTTPSServer()   // 主线程启动HTTPS服务器
}
