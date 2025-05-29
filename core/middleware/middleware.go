package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/guyigood/gyweb/core/context"
)

// HandlerFunc 使用 context 包中的 HandlerFunc 类型
type HandlerFunc = context.HandlerFunc

// Logger 日志中间件
func Logger() HandlerFunc {
	return func(c *context.Context) {
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
	return func(c *context.Context) {
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
	return func(c *context.Context) {
		c.SetHeader("Access-Control-Allow-Origin", "*")
		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.SetHeader("Access-Control-Allow-Credentials", "true")

		if c.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		c.Next()
	}
}

// Auth 认证中间件
func Auth() HandlerFunc {
	return func(c *context.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.Fail(http.StatusUnauthorized, "Unauthorized")
			return
		}
		// TODO: 实现具体的认证逻辑
		c.Next()
	}
}

// RateLimit 限流中间件
func RateLimit(limit int) HandlerFunc {
	// 使用令牌桶算法实现限流
	tokens := make(chan struct{}, limit)
	ticker := time.NewTicker(time.Second)

	// 定期补充令牌
	go func() {
		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}
	}()

	return func(c *context.Context) {
		select {
		case <-tokens:
			c.Next()
		default:
			c.Fail(http.StatusTooManyRequests, "Too Many Requests")
		}
	}
}
