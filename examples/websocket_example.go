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

// 聊天室示例
func main() {
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

	// WebSocket 配置
	wsConfig := websocket.DefaultConfig()
	wsConfig.OnConnect = func(conn *websocket.Connection) {
		log.Printf("新用户连接: %s", conn.GetClientID())

		// 向所有用户广播有新用户加入
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

		// 发送当前在线用户列表
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

	wsConfig.OnDisconnect = func(conn *websocket.Connection) {
		log.Printf("用户断开连接: %s", conn.GetClientID())

		// 向所有用户广播有用户离开
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

	wsConfig.OnMessage = func(conn *websocket.Connection, msg *websocket.Message) {
		log.Printf("收到消息: %s -> %s: %v", msg.From, msg.Type, msg.Data)

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
				Data:      "ok",
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
    <title>WebSocket 聊天室</title>
    <meta charset="UTF-8">
</head>
<body>
    <div id="messages" style="height: 400px; border: 1px solid #ccc; overflow-y: scroll; padding: 10px;"></div>
    <input type="text" id="messageInput" placeholder="输入消息..." style="width: 300px;">
    <button onclick="sendMessage()">发送</button>
    <button onclick="sendPrivateMessage()">私聊</button>
    <br><br>
    <div>在线用户: <span id="onlineUsers"></span></div>
    
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws?user_id=' + Math.random().toString(36).substr(2, 9));
        const messages = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const onlineUsers = document.getElementById('onlineUsers');
        
        ws.onopen = function() {
            addMessage('系统', '连接成功!', 'green');
        };
        
        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            console.log('收到消息:', message);
            
            switch(message.type) {
                case 'chat':
                    addMessage(message.from, message.data.content, 'black');
                    break;
                case 'private_message':
                    addMessage(message.from + '(私聊)', message.data.content, 'blue');
                    break;
                case 'user_joined':
                case 'user_left':
                    addMessage('系统', message.data.message, 'orange');
                    break;
                case 'online_users':
                    onlineUsers.textContent = message.data.users.join(', ');
                    break;
                case 'pong':
                    console.log('收到心跳响应');
                    break;
            }
        };
        
        ws.onclose = function() {
            addMessage('系统', '连接断开', 'red');
        };
        
        function addMessage(from, content, color) {
            const div = document.createElement('div');
            div.style.color = color;
            div.innerHTML = '<strong>' + from + ':</strong> ' + content + ' <small>(' + new Date().toLocaleTimeString() + ')</small>';
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }
        
        function sendMessage() {
            const content = messageInput.value.trim();
            if (content) {
                ws.send(JSON.stringify({
                    type: 'chat',
                    data: {
                        content: content
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
                        content: content
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
                ws.send(JSON.stringify({type: 'ping'}));
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

	log.Println("WebSocket 聊天室启动: http://localhost:8080")
	log.Println("API 文档:")
	log.Println("- GET  /           主页（聊天室界面）")
	log.Println("- GET  /ws         WebSocket 连接端点")
	log.Println("- GET  /api/online 获取在线用户")
	log.Println("- POST /api/broadcast 发送系统广播")

	r.Run(":8080")
}
