package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/guyigood/gyweb/core/gyarn"
)

// AuthConfig 认证中间件配置
type AuthConfig struct {
	// 认证函数，返回 true 表示认证通过
	AuthFunc func(*gyarn.Context) bool
	// 白名单路径列表（精确匹配）
	WhitelistPaths []string
	// 白名单路径前缀列表
	WhitelistPrefixes []string
	// 白名单正则表达式列表
	WhitelistPatterns []*regexp.Regexp
	// 未认证时的处理函数
	UnauthorizedHandler func(*gyarn.Context)
}

// AuthManager 认证管理器
type AuthManager struct {
	config *AuthConfig
}

// NewAuthManager 创建认证管理器
func NewAuthManager() *AuthManager {
	return &AuthManager{
		config: NewAuthConfig(),
	}
}

// UseJWT 使用JWT认证
func (m *AuthManager) UseJWT(config *JWTConfig) *AuthManager {
	m.config = NewAuthConfig()
	m.config.SetAuthFunc(func(c *gyarn.Context) bool {
		// TODO: 实现JWT验证逻辑
		return false
	})
	m.config.AddWhitelistPath(
		"/api/login",
		"/api/register",
		"/api/health",
	)
	m.config.AddWhitelistPrefix(
		"/static/",
		"/public/",
	)
	return m
}

// UseSession 使用Session认证
func (m *AuthManager) UseSession() *AuthManager {
	m.config = NewAuthConfig()
	m.config.SetAuthFunc(func(c *gyarn.Context) bool {
		// TODO: 实现Session验证逻辑
		return false
	})
	m.config.AddWhitelistPath(
		"/login",
		"/register",
		"/health",
	)
	return m
}

// UseBasic 使用Basic认证
func (m *AuthManager) UseBasic(users map[string]string) *AuthManager {
	m.config = NewAuthConfig()
	m.config.SetAuthFunc(func(c *gyarn.Context) bool {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			return false
		}
		expectedPassword, exists := users[username]
		return exists && password == expectedPassword
	})
	return m
}

// UseCustom 使用自定义认证
func (m *AuthManager) UseCustom(authFunc func(*gyarn.Context) bool) *AuthManager {
	m.config.SetAuthFunc(authFunc)
	return m
}

// AddWhitelist 添加白名单
func (m *AuthManager) AddWhitelist(paths []string, prefixes []string, patterns []string) *AuthManager {
	if paths != nil {
		m.config.AddWhitelistPath(paths...)
	}
	if prefixes != nil {
		m.config.AddWhitelistPrefix(prefixes...)
	}
	if patterns != nil {
		m.config.AddWhitelistPattern(patterns...)
	}
	return m
}

// SetUnauthorizedHandler 设置未授权处理函数
func (m *AuthManager) SetUnauthorizedHandler(handler func(*gyarn.Context)) *AuthManager {
	m.config.SetUnauthorizedHandler(handler)
	return m
}

// Build 构建认证中间件
func (m *AuthManager) Build() gyarn.HandlerFunc {
	return CreateAuthMiddleware(m.config)
}

// CreateAuthMiddleware 创建认证中间件
func CreateAuthMiddleware(config *AuthConfig) gyarn.HandlerFunc {
	if config == nil {
		config = NewAuthConfig()
	}

	// 设置默认的未认证处理函数
	if config.UnauthorizedHandler == nil {
		config.UnauthorizedHandler = func(c *gyarn.Context) {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "未授权访问",
			})
		}
	}

	return func(c *gyarn.Context) {
		debugAuth(c, "开始认证检查")

		// 检查是否在白名单中
		if isWhitelisted(c.Path, config) {
			debugWhitelist(c.Path, true, "whitelist")
			c.Next()
			return
		}

		// 执行认证
		if config.AuthFunc != nil {
			success := config.AuthFunc(c)
			debugAuthFunc(c, success)
			if success {
				c.Next()
				return
			}
		}

		// 认证失败
		debugUnauthorized(c)
		config.UnauthorizedHandler(c)
		c.Abort()
	}
}

// isWhitelisted 检查路径是否在白名单中
func isWhitelisted(path string, config *AuthConfig) bool {
	// 检查精确匹配
	for _, p := range config.WhitelistPaths {
		if path == p {
			debugWhitelist(path, true, "exact")
			return true
		}
	}

	// 检查前缀匹配
	for _, prefix := range config.WhitelistPrefixes {
		if strings.HasPrefix(path, prefix) {
			debugWhitelist(path, true, "prefix")
			return true
		}
	}

	// 检查正则匹配
	for _, pattern := range config.WhitelistPatterns {
		if pattern.MatchString(path) {
			debugWhitelist(path, true, "regex")
			return true
		}
	}

	debugWhitelist(path, false, "none")
	return false
}

// NewAuthConfig 创建认证配置
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		WhitelistPaths:      make([]string, 0),
		WhitelistPrefixes:   make([]string, 0),
		WhitelistPatterns:   make([]*regexp.Regexp, 0),
		UnauthorizedHandler: nil,
	}
}

// AddWhitelistPath 添加白名单路径
func (c *AuthConfig) AddWhitelistPath(paths ...string) *AuthConfig {
	c.WhitelistPaths = append(c.WhitelistPaths, paths...)
	return c
}

// AddWhitelistPrefix 添加白名单路径前缀
func (c *AuthConfig) AddWhitelistPrefix(prefixes ...string) *AuthConfig {
	c.WhitelistPrefixes = append(c.WhitelistPrefixes, prefixes...)
	return c
}

// AddWhitelistPattern 添加白名单正则表达式
func (c *AuthConfig) AddWhitelistPattern(patterns ...string) *AuthConfig {
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			c.WhitelistPatterns = append(c.WhitelistPatterns, regex)
		}
	}
	return c
}

// SetAuthFunc 设置认证函数
func (c *AuthConfig) SetAuthFunc(fn func(*gyarn.Context) bool) *AuthConfig {
	c.AuthFunc = fn
	return c
}

// SetUnauthorizedHandler 设置未认证处理函数
func (c *AuthConfig) SetUnauthorizedHandler(handler func(*gyarn.Context)) *AuthConfig {
	c.UnauthorizedHandler = handler
	return c
}

// 一些常用的认证配置预设

// JWTConfig JWT认证配置
type JWTConfig struct {
	SecretKey     string
	TokenLookup   string // 格式: "header:Authorization" 或 "query:token" 或 "cookie:token"
	TokenHeadName string // 例如: "Bearer"
}

// CreateJWTAuth 创建JWT认证中间件
func CreateJWTAuth(config *JWTConfig) gyarn.HandlerFunc {
	authConfig := NewAuthConfig()

	// 设置JWT认证函数
	authConfig.SetAuthFunc(func(c *gyarn.Context) bool {
		// TODO: 实现JWT验证逻辑
		// 这里需要根据实际使用的JWT库来实现
		return false
	})

	// 添加一些默认的白名单路径
	authConfig.AddWhitelistPath(
		"/api/login",
		"/api/register",
		"/api/health",
	)

	// 添加静态文件前缀
	authConfig.AddWhitelistPrefix(
		"/static/",
		"/public/",
	)

	return CreateAuthMiddleware(authConfig)
}

// CreateSessionAuth 创建Session认证中间件
func CreateSessionAuth() gyarn.HandlerFunc {
	authConfig := NewAuthConfig()

	// 设置Session认证函数
	authConfig.SetAuthFunc(func(c *gyarn.Context) bool {
		// TODO: 实现Session验证逻辑
		// 这里需要根据实际使用的Session库来实现
		return false
	})

	// 添加一些默认的白名单路径
	authConfig.AddWhitelistPath(
		"/login",
		"/register",
		"/health",
	)

	return CreateAuthMiddleware(authConfig)
}

// CreateBasicAuth 创建Basic认证中间件
func CreateBasicAuth(users map[string]string) gyarn.HandlerFunc {
	authConfig := NewAuthConfig()

	// 设置Basic认证函数
	authConfig.SetAuthFunc(func(c *gyarn.Context) bool {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			return false
		}
		expectedPassword, exists := users[username]
		return exists && password == expectedPassword
	})

	return CreateAuthMiddleware(authConfig)
}
