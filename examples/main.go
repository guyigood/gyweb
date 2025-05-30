package main

import (
	"net/http"
	"time"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

func main() {
	// 创建引擎实例
	r := engine.New()
	middleware.SetDebug(true)

	// 只使用基础中间件，不使用认证中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 注册路由
	r.GET("/", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Welcome to GyWeb!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 健康检查路由
	r.GET("/api/health", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 路由组
	api := r.Group("/api")
	{
		// 用户相关路由（公开访问）
		users := api.Group("/users")
		{
			users.GET("", func(c *gyarn.Context) {
				c.JSON(http.StatusOK, gyarn.H{
					"users": []string{"Alice", "Bob", "Charlie"},
				})
			})

			users.GET("/:id", func(c *gyarn.Context) {
				c.JSON(http.StatusOK, gyarn.H{
					"id":   c.Param("id"),
					"name": "User " + c.Param("id"),
				})
			})

			users.POST("", func(c *gyarn.Context) {
				name := c.PostForm("name")
				c.JSON(http.StatusCreated, gyarn.H{
					"message": "User created",
					"name":    name,
				})
			})
		}

		// 需要认证的管理路由
		admin := api.Group("/admin")
		// 只在这个路由组上使用认证中间件
		authMiddleware := middleware.NewAuthManager().
			UseCustom(func(c *gyarn.Context) bool {
				// 简单的认证逻辑：检查是否有Authorization头部
				token := c.Request.Header.Get("Authorization")
				return token != ""
			}).
			Build()
		admin.Use(authMiddleware)
		{
			admin.GET("/dashboard", func(c *gyarn.Context) {
				c.JSON(http.StatusOK, gyarn.H{
					"message": "Welcome to admin dashboard",
				})
			})
		}
	}

	// 启动服务器
	r.Run(":8080")
}
