package lib

import (
	"io"
	"time"
	"{project_name}/model"
	"{project_name}/public"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

func LogDb() middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		t := time.Now().Format("2006-01-02 15:04:05")
		db := public.Db
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
		}

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
		c.Next()
	}

}
