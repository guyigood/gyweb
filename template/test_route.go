package main

import (
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
)

// AddTestRoutes 添加测试路由
func AddTestRoutes(r *engine.Engine) {
	// 测试路由组
	test := r.Group("/api/test")
	{
		// 无需认证的测试路由
		test.GET("", func(c *gyarn.Context) {
			c.JSON(http.StatusOK, gyarn.H{
				"message": "GET请求成功",
				"method":  c.Method,
				"path":    c.Path,
			})
		})

		test.POST("", func(c *gyarn.Context) {
			c.JSON(http.StatusOK, gyarn.H{
				"message": "POST请求成功",
				"method":  c.Method,
				"path":    c.Path,
			})
		})

		// 需要认证的测试路由
		test.POST("/protected", func(c *gyarn.Context) {
			c.JSON(http.StatusOK, gyarn.H{
				"message": "受保护的POST请求成功",
				"method":  c.Method,
				"path":    c.Path,
			})
		})
	}

	// 调试路由 - 显示请求详情
	r.GET("/debug/request", func(c *gyarn.Context) {
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		c.JSON(http.StatusOK, gyarn.H{
			"method":  c.Method,
			"path":    c.Path,
			"headers": headers,
			"query":   c.Request.URL.Query(),
		})
	})

	// 专门测试CORS的路由
	r.OPTIONS("/api/test", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Manual OPTIONS handler",
		})
	})
}
