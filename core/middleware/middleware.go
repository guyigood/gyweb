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

		// 设置CORS头部 - 修复Origin和Credentials冲突问题
		if origin != "" {
			// 有Origin时，使用具体的Origin值
			c.SetHeader("Access-Control-Allow-Origin", origin)
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		} else {
			// 没有Origin时（同域请求），可以使用*
			c.SetHeader("Access-Control-Allow-Origin", "*")
			// 注意：设置*时不能设置Credentials为true
		}

		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Custom-Header, Token")
		c.SetHeader("Access-Control-Max-Age", "86400") // 24小时缓存预检结果

		// 对于OPTIONS请求，立即响应，不继续执行后续中间件
		if c.Method == "OPTIONS" {
			log.Printf("[CORS] OPTIONS预检请求，立即响应")

			// 检查预检请求的头部
			requestMethod := c.GetHeader("Access-Control-Request-Method")
			requestHeaders := c.GetHeader("Access-Control-Request-Headers")

			log.Printf("[CORS] 预检请求 - Method: %s, Headers: %s", requestMethod, requestHeaders)

			// 添加暴露的头部
			c.SetHeader("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")

			// 设置正确的Content-Type和状态码
			c.SetHeader("Content-Type", "text/plain; charset=utf-8")
			c.Status(http.StatusOK) // 使用200状态码

			// 写入空响应体（重要：必须写入响应体，即使是空的）
			c.Writer.Write([]byte(""))

			log.Printf("[CORS] OPTIONS响应已发送，中止后续中间件执行")
			c.Abort()
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

// RateLimit 限流中间件 - 简化版本，向后兼容
func RateLimit(limit int) HandlerFunc {
	return RateLimitAdvanced(RateLimitConfig{
		RequestsPerSecond: limit,
		BurstSize:         limit,
		Debug:             false,
	})
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每秒允许的请求数
	RequestsPerSecond int
	// 突发请求数（令牌桶容量）
	BurstSize int
	// 是否启用调试日志
	Debug bool
}

// RateLimitAdvanced 高级限流中间件，支持更多配置
func RateLimitAdvanced(config RateLimitConfig) HandlerFunc {
	if config.RequestsPerSecond <= 0 {
		config.RequestsPerSecond = 100
	}
	if config.BurstSize <= 0 {
		config.BurstSize = config.RequestsPerSecond
	}

	// 使用令牌桶算法
	tokens := make(chan struct{}, config.BurstSize)

	// 初始化时填充所有令牌
	for i := 0; i < config.BurstSize; i++ {
		tokens <- struct{}{}
	}

	// 计算补充间隔
	interval := time.Second / time.Duration(config.RequestsPerSecond)
	if config.Debug {
		log.Printf("[RateLimit] 配置: %d req/s, 突发: %d, 补充间隔: %v",
			config.RequestsPerSecond, config.BurstSize, interval)
	}

	// 定期补充令牌
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			select {
			case tokens <- struct{}{}:
				// 成功添加令牌
			default:
				// 令牌桶已满
			}
		}
	}()

	return func(c *gyarn.Context) {
		select {
		case <-tokens:
			if config.Debug {
				log.Printf("[RateLimit] 请求通过: %s %s", c.Method, c.Path)
			}
			c.Next()
		default:
			if config.Debug {
				log.Printf("[RateLimit] 请求被限流: %s %s", c.Method, c.Path)
			}
			c.Fail(http.StatusTooManyRequests, "Too Many Requests")
		}
	}
}
