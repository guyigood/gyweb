package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
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

// JWTConfig JWT认证配置
type JWTConfig struct {
	SecretKey     string        // JWT密钥
	TokenLookup   string        // 格式: "header:Authorization" 或 "query:token" 或 "cookie:token"
	TokenHeadName string        // 例如: "Bearer"
	ExpiresIn     time.Duration // token过期时间
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// SessionConfig Session认证配置
type SessionConfig struct {
	SecretKey string // Session密钥
	MaxAge    int    // Session过期时间（秒）
	Path      string // Cookie路径
	Domain    string // Cookie域名
	Secure    bool   // 是否只在HTTPS下传输
	HttpOnly  bool   // 是否禁止JavaScript访问
}

// SessionStore Session存储
var sessionStore *sessions.CookieStore

// InitSessionStore 初始化Session存储
func InitSessionStore(config *SessionConfig) {
	// 生成随机密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	if config.SecretKey != "" {
		key = []byte(config.SecretKey)
	}
	sessionStore = sessions.NewCookieStore(key)
}

// UseJWT 使用JWT认证
func (m *AuthManager) UseJWT(config *JWTConfig) *AuthManager {
	m.config = NewAuthConfig()

	// 设置默认值
	if config.ExpiresIn == 0 {
		config.ExpiresIn = 24 * time.Hour
	}
	if config.TokenHeadName == "" {
		config.TokenHeadName = "Bearer"
	}
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}

	m.config.SetAuthFunc(func(c *gyarn.Context) bool {
		// 从请求中获取token
		var tokenString string
		parts := strings.Split(config.TokenLookup, ":")
		if len(parts) != 2 {
			return false
		}

		switch parts[0] {
		case "header":
			tokenString = c.GetHeader(parts[1])
		case "query":
			tokenString = c.Query(parts[1])
		case "cookie":
			tokenString, _ = c.Cookie(parts[1])
		default:
			return false
		}

		// 移除token前缀
		if config.TokenHeadName != "" && strings.HasPrefix(tokenString, config.TokenHeadName+" ") {
			tokenString = strings.TrimPrefix(tokenString, config.TokenHeadName+" ")
		}

		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.SecretKey), nil
		})

		if err != nil || !token.Valid {
			return false
		}

		// 将用户信息存储到上下文中
		if claims, ok := token.Claims.(*JWTClaims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			return true
		}

		return false
	})

	// 添加默认白名单
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
func (m *AuthManager) UseSession(config *SessionConfig) *AuthManager {
	m.config = NewAuthConfig()

	// 初始化Session存储
	if sessionStore == nil {
		InitSessionStore(config)
	}

	m.config.SetAuthFunc(func(c *gyarn.Context) bool {
		// 获取session
		session, err := sessionStore.Get(c.Request, "session")
		if err != nil {
			return false
		}

		// 检查用户是否已登录
		userID, ok := session.Values["user_id"]
		if !ok {
			return false
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", userID)
		if username, ok := session.Values["username"]; ok {
			c.Set("username", username)
		}

		return true
	})

	// 添加默认白名单
	m.config.AddWhitelistPath(
		"/login",
		"/register",
		"/health",
	)

	return m
}

// GenerateJWT 生成JWT token
func GenerateJWT(config *JWTConfig, userID string, username string, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// SetSession 设置Session
func SetSession(c *gyarn.Context, config *SessionConfig, userID int64, username, role string) error {
	session, err := sessionStore.Get(c.Request, "session")
	if err != nil {
		return err
	}

	session.Values["user_id"] = userID
	session.Values["username"] = username
	session.Values["role"] = role
	session.Options = &sessions.Options{
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   config.MaxAge,
		Secure:   config.Secure,
		HttpOnly: config.HttpOnly,
	}

	return session.Save(c.Request, c.Writer)
}

// ClearSession 清除Session
func ClearSession(c *gyarn.Context) error {
	session, err := sessionStore.Get(c.Request, "session")
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	return session.Save(c.Request, c.Writer)
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

		// 如果没有设置认证函数，直接通过
		if config.AuthFunc == nil {
			debugAuth(c, "未设置认证函数，直接通过")
			c.Next()
			return
		}

		// 执行认证
		success := config.AuthFunc(c)
		debugAuthFunc(c, success)
		if success {
			c.Next()
			return
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
