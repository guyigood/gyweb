package lib

import (
	"time"
	"{firstweb}/model"
	"{firstweb}/public"

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
		db.Table("sl_log").Insert(gyarn.H{
			"ip":       c.ClientIP(),
			"url":      c.Request.URL.Path,
			"add_time": t,
			"user_id":  user_id,
			"method":   c.Request.Method,
			"params":   c.Request.URL.Query(),
			"body":     c.Request.Body,
		})
		c.Next()
	}

}
