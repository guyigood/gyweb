package main

import (
	"log"
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

func main() {
	// 创建引擎实例
	r := engine.New()
	middleware.SetDebug(true)

	// 使用基础中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 注册基础路由
	r.GET("/", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Welcome to GyWeb!",
		})
	})

	r.GET("/health", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"status": "ok",
		})
	})

	log.Println("Server starting on :8084")
	r.Run(":8084")
}
