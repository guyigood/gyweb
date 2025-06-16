package lib

import (
	"net/http"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

// OptionsHandler 处理OPTIONS预检请求
func OptionsHandler() middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		// 如果是OPTIONS请求，直接返回成功响应
		if c.Method == "OPTIONS" {
			// 设置CORS头部
			c.SetHeader("Access-Control-Allow-Origin", "*")
			c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.SetHeader("Access-Control-Allow-Credentials", "true")
			c.SetHeader("Access-Control-Max-Age", "86400") // 24小时

			// 立即返回成功状态，不继续执行后续中间件
			c.Status(http.StatusOK)
			c.Abort()
			return
		}

		// 不是OPTIONS请求，继续处理
		c.Next()
	}
}
