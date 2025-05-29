package main

import (
	"net/http"
	"time"

	"github.com/guyigood/gyweb/core/context"
	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/middleware"
)

func main() {
	// 创建引擎实例
	r := engine.New()
	middleware.SetDebug(true)
	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Auth())
	r.Use(middleware.RateLimit(100))
	r.Use(middleware.CreateSessionAuth())

	// 注册路由
	r.GET("/", func(c *context.Context) {
		c.JSON(http.StatusOK, context.H{
			"message": "Welcome to GyWeb!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		{
			users.GET("", func(c *context.Context) {

				c.JSON(http.StatusOK, context.H{
					"users": []string{"Alice", "Bob", "Charlie"},
				})
			})

			users.GET("/:id", func(c *context.Context) {
				c.JSON(http.StatusOK, context.H{
					"id":   c.Param("id"),
					"name": "User " + c.Param("id"),
				})
			})

			users.POST("", func(c *context.Context) {
				name := c.PostForm("name")
				c.JSON(http.StatusCreated, context.H{
					"message": "User created",
					"name":    name,
				})
			})
		}

		// 需要认证的路由
		admin := api.Group("/admin")
		admin.Use(middleware.Auth())
		{

			admin.GET("/dashboard", func(c *context.Context) {
				c.JSON(http.StatusOK, context.H{
					"message": "Welcome to admin dashboard",
				})
			})
		}
	}

	// 启动服务器
	r.Run(":8080")
}
