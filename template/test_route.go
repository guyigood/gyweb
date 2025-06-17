package main

import (
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
)

// AddTestRoutes 添加测试路由
func AddTestRoutes(r *engine.Engine) {
	// 测试路由组 - 确保在认证白名单中
	test := r.Group("/api/test")
	{
		// 无需认证的测试路由
		test.GET("", func(c *gyarn.Context) {
			c.JSON(http.StatusOK, gyarn.H{
				"message": "GET请求成功",
				"method":  c.Method,
				"path":    c.Path,
				"headers": map[string]string{
					"Origin":       c.GetHeader("Origin"),
					"Content-Type": c.GetHeader("Content-Type"),
					"User-Agent":   c.GetHeader("User-Agent"),
				},
			})
		})

		test.POST("", func(c *gyarn.Context) {
			// 读取请求体
			var body map[string]interface{}
			c.BindJSON(&body)

			c.JSON(http.StatusOK, gyarn.H{
				"message": "POST请求成功！🎉",
				"method":  c.Method,
				"path":    c.Path,
				"body":    body,
				"headers": map[string]string{
					"Origin":        c.GetHeader("Origin"),
					"Content-Type":  c.GetHeader("Content-Type"),
					"Authorization": c.GetHeader("Authorization"),
				},
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
			"message":   "调试信息",
			"method":    c.Method,
			"path":      c.Path,
			"headers":   headers,
			"query":     c.Request.URL.Query(),
			"client_ip": c.ClientIP(),
		})
	})

	// CORS状态检查路由
	r.GET("/cors/status", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"message": "CORS配置正常",
			"cors_headers": map[string]string{
				"Access-Control-Allow-Origin":      c.Writer.Header().Get("Access-Control-Allow-Origin"),
				"Access-Control-Allow-Methods":     c.Writer.Header().Get("Access-Control-Allow-Methods"),
				"Access-Control-Allow-Headers":     c.Writer.Header().Get("Access-Control-Allow-Headers"),
				"Access-Control-Allow-Credentials": c.Writer.Header().Get("Access-Control-Allow-Credentials"),
			},
		})
	})
}
