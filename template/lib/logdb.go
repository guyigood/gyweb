package lib

import (
	"bytes"
	"io"
	"{project_name}/model"
	"{project_name}/public"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

func LogDb() middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		// 跳过OPTIONS请求的日志记录，避免CORS预检请求时的中间件冲突
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		t := time.Now().Format("2006-01-02 15:04:05")
		db := public.GetDb()
		login, ok := c.Get("login")
		user_id := 0
		if ok {
			user, u_flag := login.(model.LoginUser)
			if u_flag {
				user_id = user.ID
			}
		}
		body_str := ""
		if c.Request.Body != nil {
			body := c.Request.Body
			defer body.Close()
			body_b, _ := io.ReadAll(body)
			body_str = string(body_b)
			// 重新设置请求体，以便后续处理函数可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body_b))
		}

		// 异步记录日志，避免阻塞请求处理
		go func() {
			_, err := db.Table("operation_log").Insert(map[string]interface{}{
				"ip":       c.ClientIP(),
				"url":      c.Request.URL.Path,
				"add_time": t,
				"user_id":  user_id,
				"method":   c.Request.Method,
				"params":   c.Request.URL.Query().Encode(),
				"body":     body_str,
			})
			middleware.DebugVar("log", err)
		}()

		c.Next()
	}

}
