package middleware

import (
	"fmt"
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
		c.SetHeader("Access-Control-Allow-Origin", "*")
		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.SetHeader("Access-Control-Allow-Credentials", "true")

		if c.Method == "OPTIONS" {
			fmt.Println("CORS OPTIONS")
			c.Status(http.StatusOK)
			c.Abort() // 停止后续中间件的执行
			return
		}

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
