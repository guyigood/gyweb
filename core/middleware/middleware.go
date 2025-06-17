package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
)

// HandlerFunc 使用 context 包中的 HandlerFunc 类型
type HandlerFunc = gyarn.HandlerFunc

// Logger 日志中间件
func Logger() HandlerFunc {
	return func(c *gyarn.Context) {
		// 开始时间
		t := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		log.Printf("[%d] %s in %v", c.StatusCode, c.Request.URL.Path, time.Since(t))
	}
}

// Recovery 恢复中间件
func Recovery() HandlerFunc {
	return func(c *gyarn.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() HandlerFunc {
	return func(c *gyarn.Context) {
		origin := c.GetHeader("Origin")

		// 添加调试日志
		log.Printf("[CORS] 处理请求: %s %s, Origin: %s", c.Method, c.Path, origin)

		// 设置CORS头部 - 更完整的配置
		if origin != "" {
			c.SetHeader("Access-Control-Allow-Origin", origin) // 使用具体的Origin而不是*
		} else {
			c.SetHeader("Access-Control-Allow-Origin", "*")
		}

		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Custom-Header")
		c.SetHeader("Access-Control-Allow-Credentials", "true")
		c.SetHeader("Access-Control-Max-Age", "86400") // 24小时缓存预检结果

		// 对于OPTIONS请求，还需要设置额外的头部
		if c.Method == "OPTIONS" {
			log.Printf("[CORS] OPTIONS请求，设置完整的预检响应头")

			// 添加暴露的头部
			c.SetHeader("Access-Control-Expose-Headers", "Content-Length, Content-Type")

			// 确保Content-Type也被设置
			c.SetHeader("Content-Type", "text/plain; charset=utf-8")
			c.SetHeader("Content-Length", "0")

			// 立即返回成功状态，不继续执行后续中间件
			c.Status(http.StatusOK)
			c.Abort() // 关键：停止后续中间件执行，防止进入认证中间件
			return
		}

		log.Printf("[CORS] 非OPTIONS请求，继续执行后续中间件")
		c.Next()
	}
}

// Auth 认证中间件（已废弃，请使用 NewAuthManager）
// Deprecated: 请使用 NewAuthManager 创建认证中间件
func Auth() HandlerFunc {
	return NewAuthManager().
		UseCustom(func(c *gyarn.Context) bool {
			token := c.Request.Header.Get("Authorization")
			return token != ""
		}).
		Build()
}

// RateLimit 限流中间件
func RateLimit(limit int) HandlerFunc {
	// 使用令牌桶算法实现限流
	tokens := make(chan struct{}, limit)
	ticker := time.NewTicker(time.Second)
	done := make(chan struct{})

	// 定期补充令牌
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
				}
			case <-done:
				return
			}
		}
	}()

	return func(c *gyarn.Context) {
		select {
		case <-tokens:
			c.Next()
		default:
			c.Fail(http.StatusTooManyRequests, "Too Many Requests")
		}
		// 当中间件被销毁时，停止 ticker
		if c.IsAborted() {
			close(done)
		}
	}
}
