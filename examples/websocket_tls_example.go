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

// WSS 聊天室示例
func mainTLS() {
	r := engine.New()

	// 启用调试模式
	middleware.SetDebug(true)

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 创建 WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 设置安全的来源验证
	websocket.SetSecureCheckOrigin()

	// WebSocket TLS 配置
	wsConfig := websocket.DefaultTLSConfig()
	wsConfig.OnConnect = func(conn *websocket.Connection) {
		log.Printf("新用户连接 (WSS): %s", conn.GetClientID())

		// 向所有用户广播有新用户加入
		message := websocket.Message{
			Type: "user_joined",
			Data: map[string]interface{}{
				"user_id": conn.GetClientID(),
				"message": fmt.Sprintf("用户 %s 加入了安全聊天室", conn.GetClientID()),
				"secure":  true,
			},
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
		}

		// 发送当前在线用户列表
		onlineUsers := wsHub.GetConnections()
		userListMessage := websocket.Message{
			Type: "online_users",
			Data: map[string]interface{}{
				"users":  onlineUsers,
				"count":  len(onlineUsers),
				"secure": true,
			},
			To:        conn.GetClientID(),
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(userListMessage); err == nil {
			conn.SendMessage(data)
		}
	}

	wsConfig.OnDisconnect = func(conn *websocket.Connection) {
		log.Printf("用户断开连接 (WSS): %s", conn.GetClientID())

		// 向所有用户广播有用户离开
		message := websocket.Message{
			Type: "user_left",
			Data: map[string]interface{}{
				"user_id": conn.GetClientID(),
				"message": fmt.Sprintf("用户 %s 离开了安全聊天室", conn.GetClientID()),
				"secure":  true,
			},
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
		}
	}

	wsConfig.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
		log.Printf("收到安全消息: %s -> %s: %v", msg.From, msg.Type, msg.Data)

		switch msg.Type {
		case "chat":
			// 聊天消息，广播给所有用户
			if data, err := json.Marshal(msg); err == nil {
				wsHub.Broadcast(data)
			}

		case "private_message":
			// 私聊消息，只发送给指定用户
			if msg.To != "" {
				if data, err := json.Marshal(msg); err == nil {
					wsHub.SendToClient(msg.To, data)
				}
			}

		case "ping":
			// 心跳检测
			pongMessage := websocket.Message{
				Type:      "pong",
				Data:      "secure_ok",
				To:        conn.GetClientID(),
				Timestamp: time.Now().Unix(),
			}
			if data, err := json.Marshal(pongMessage); err == nil {
				conn.SendMessage(data)
			}
		}
	}

	// 主页
	r.GET("/", func(c *gyarn.Context) {
		c.HTML(200, `
<!DOCTYPE html>
<html>
<head>
    <title>WSS 安全聊天室</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .secure-badge { 
            background: #28a745; 
            color: white; 
            padding: 2px 8px; 
            border-radius: 4px; 
            font-size: 12px; 
            margin-left: 10px; 
        }
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
    </style>
</head>
<body>
    <div class="container">
        <h1>WSS 安全聊天室 <span class="secure-badge">🔒 安全连接</span></h1>
        
        <div id="status" class="status">连接中...</div>
        
        <div id="messages"></div>
        
        <div class="input-group">
            <input type="text" id="messageInput" placeholder="输入消息...">
            <button class="btn-primary" onclick="sendMessage()">发送</button>
            <button class="btn-secondary" onclick="sendPrivateMessage()">私聊</button>
        </div>
        
        <div>在线用户 <span class="secure-badge">🔒</span>: <span id="onlineUsers">0</span></div>
    </div>
    
    <script>
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = protocol + '//' + window.location.host + '/ws?user_id=' + Math.random().toString(36).substr(2, 9);
        const ws = new WebSocket(wsUrl);
        
        const messages = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const onlineUsers = document.getElementById('onlineUsers');
        const status = document.getElementById('status');
        
        ws.onopen = function() {
            addMessage('系统', '🔒 安全连接建立成功!', 'green');
            status.textContent = '🔒 已连接 (WSS)';
            status.className = 'status connected';
        };
        
        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            console.log('收到安全消息:', message);
            
            switch(message.type) {
                case 'chat':
                    addMessage(message.from + ' 🔒', message.data.content, 'black');
                    break;
                case 'private_message':
                    addMessage(message.from + ' 🔒 (私聊)', message.data.content, 'blue');
                    break;
                case 'user_joined':
                case 'user_left':
                    addMessage('🔒 系统', message.data.message, 'orange');
                    break;
                case 'online_users':
                    onlineUsers.textContent = message.data.count + ' 人在线 (安全连接)';
                    break;
                case 'pong':
                    console.log('收到安全心跳响应:', message.data);
                    break;
            }
        };
        
        ws.onclose = function() {
            addMessage('系统', '🔒 安全连接断开', 'red');
            status.textContent = '🔒 连接断开';
            status.className = 'status disconnected';
        };
        
        ws.onerror = function(error) {
            addMessage('系统', '🔒 连接错误: ' + error, 'red');
            status.textContent = '🔒 连接错误';
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
                        secure: true
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
                        secure: true
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
        
        // 定时发送安全心跳
        setInterval(function() {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({type: 'ping', data: {secure: true}}));
            }
        }, 30000);
    </script>
</body>
</html>
		`)
	})

	// WebSocket 端点
	r.GET("/ws", func(c *gyarn.Context) {
		// 从查询参数获取用户ID，也可以从认证中间件获取
		userID := c.Query("user_id")
		if userID == "" {
			userID = "user_" + strconv.FormatInt(time.Now().UnixNano(), 36)
		}

		// 处理 WebSocket 连接
		websocket.HandleWebSocket(c, wsHub, wsConfig, userID)
	})

	// API 端点 - 获取在线用户
	r.GET("/api/online", func(c *gyarn.Context) {
		users := wsHub.GetConnections()
		c.Success(gyarn.H{
			"users":  users,
			"count":  len(users),
			"secure": true,
		})
	})

	// API 端点 - 发送安全广播消息
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
				"secure":  true,
			},
			From:      "system",
			Timestamp: time.Now().Unix(),
		}

		if data, err := json.Marshal(message); err == nil {
			wsHub.Broadcast(data)
			c.Success(gyarn.H{"message": "安全广播发送成功"})
		} else {
			c.InternalServerError("发送失败")
		}
	})

	log.Println("WSS 安全聊天室启动选项:")
	log.Println("1. 使用自签名证书 (开发环境): https://localhost:8443")
	log.Println("2. 使用自定义证书 (生产环境): 配置 cert.pem 和 key.pem")
	log.Println()
	log.Println("API 文档:")
	log.Println("- GET  /           主页（安全聊天室界面）")
	log.Println("- GET  /ws         WSS 连接端点")
	log.Println("- GET  /api/online 获取在线用户")
	log.Println("- POST /api/broadcast 发送系统广播")

	// 启动 HTTPS 服务器
	// 方式1: 使用自签名证书（开发环境）
	if err := r.RunAutoTLS(":8443"); err != nil {
		log.Fatal("启动 HTTPS 服务器失败:", err)
	}

	// 方式2: 使用自定义证书（生产环境）
	// tlsConfig := &engine.TLSConfig{
	//     CertFile: "cert.pem",
	//     KeyFile:  "key.pem",
	// }
	// if err := r.RunTLS(":8443", tlsConfig); err != nil {
	//     log.Fatal("启动 HTTPS 服务器失败:", err)
	// }
}
