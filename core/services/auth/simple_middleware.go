package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

// SimpleAuthMiddleware 简化的鉴权中间件
type SimpleAuthMiddleware struct {
	authService *AuthService
	skipPaths   []string
	skipPrefix  []string
}

// NewSimpleAuthMiddleware 创建简化的鉴权中间件
func NewSimpleAuthMiddleware(authService *AuthService) *SimpleAuthMiddleware {
	return &SimpleAuthMiddleware{
		authService: authService,
		skipPaths:   []string{"/health", "/metrics", "/api/login", "/api/register"},
		skipPrefix:  []string{"/static/", "/public/"},
	}
}

// SkipPaths 设置跳过鉴权的路径
func (m *SimpleAuthMiddleware) SkipPaths(paths ...string) *SimpleAuthMiddleware {
	m.skipPaths = append(m.skipPaths, paths...)
	return m
}

// SkipPrefix 设置跳过鉴权的路径前缀
func (m *SimpleAuthMiddleware) SkipPrefix(prefixes ...string) *SimpleAuthMiddleware {
	m.skipPrefix = append(m.skipPrefix, prefixes...)
	return m
}

// Handler 中间件处理函数
func (m *SimpleAuthMiddleware) Handler() gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查是否跳过鉴权
		if m.shouldSkip(path) {
			c.Next()
			return
		}

		// 从上下文获取用户信息（需要在认证中间件之后使用）
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "未认证，请先登录",
			})
			c.Abort()
			return
		}

		// 如果没有配置鉴权服务，只检查认证
		if m.authService == nil {
			c.Next()
			return
		}

		// 构建鉴权上下文
		username := ""
		if usernameValue, exists := c.Get("username"); exists {
			if str, ok := usernameValue.(string); ok {
				username = str
			}
		}

		// 解析资源和操作
		resource, action := m.parseResourceAction(path, method)

		authCtx := &AuthContext{
			UserID:      userID.(string),
			Username:    username,
			RequestPath: path,
			Method:      method,
			Resource:    resource,
			Action:      action,
			Attributes:  make(map[string]string),
		}

		// 执行权限验证
		allowed, err := m.authService.Authorize(authCtx)
		if err != nil {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "权限验证失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "权限不足，无法访问该资源",
			})
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("auth_context", authCtx)
		c.Next()
	}
}

// shouldSkip 检查是否应该跳过鉴权
func (m *SimpleAuthMiddleware) shouldSkip(path string) bool {
	// 检查精确匹配
	for _, skipPath := range m.skipPaths {
		if path == skipPath {
			return true
		}
	}

	// 检查前缀匹配
	for _, prefix := range m.skipPrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// parseResourceAction 解析资源和操作
func (m *SimpleAuthMiddleware) parseResourceAction(path, method string) (string, string) {
	// 简单的解析逻辑
	parts := strings.Split(strings.Trim(path, "/"), "/")

	var resource, action string

	// 解析资源 (例如: /api/users/123 -> resource=user)
	if len(parts) >= 2 && parts[0] == "api" {
		resource = strings.TrimSuffix(parts[1], "s") // 去掉复数s
	}

	// 解析操作 (根据HTTP方法)
	switch method {
	case "GET":
		action = "read"
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "unknown"
	}

	return resource, action
}

// CreateAuthChain 创建完整的认证鉴权链
func CreateAuthChain(jwtConfig *middleware.JWTConfig, authService *AuthService) []gyarn.HandlerFunc {
	middlewares := make([]gyarn.HandlerFunc, 0)

	// 1. 先添加认证中间件
	if jwtConfig != nil {
		authMiddleware := middleware.CreateJWTAuth(jwtConfig)
		middlewares = append(middlewares, authMiddleware)
	}

	// 2. 再添加鉴权中间件
	if authService != nil {
		authzMiddleware := NewSimpleAuthMiddleware(authService).Handler()
		middlewares = append(middlewares, authzMiddleware)
	}

	return middlewares
}

// CreateSessionAuthChain 创建Session认证鉴权链
func CreateSessionAuthChain(sessionConfig *middleware.SessionConfig, authService *AuthService) []gyarn.HandlerFunc {
	middlewares := make([]gyarn.HandlerFunc, 0)

	// 1. 先添加Session认证中间件
	if sessionConfig != nil {
		middleware.InitSessionStore(sessionConfig)
		authMiddleware := middleware.CreateSessionAuth()
		middlewares = append(middlewares, authMiddleware)
	}

	// 2. 再添加鉴权中间件
	if authService != nil {
		authzMiddleware := NewSimpleAuthMiddleware(authService).Handler()
		middlewares = append(middlewares, authzMiddleware)
	}

	return middlewares
}

// RequirePermission 要求特定权限的装饰器
func RequirePermission(resource, action string) gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		authCtx, exists := c.Get("auth_context")
		if !exists {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "未找到权限上下文",
			})
			c.Abort()
			return
		}

		ctx := authCtx.(*AuthContext)

		// 检查是否有指定权限
		hasPermission := false
		for _, perm := range ctx.Permissions {
			if (perm.Resource == resource || perm.Resource == "*") &&
				(perm.Action == action || perm.Action == "*") {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": fmt.Sprintf("需要权限: %s.%s", resource, action),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求特定角色的装饰器
func RequireRole(roles ...string) gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		authCtx, exists := c.Get("auth_context")
		if !exists {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "未找到权限上下文",
			})
			c.Abort()
			return
		}

		ctx := authCtx.(*AuthContext)

		// 检查是否有指定角色
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range ctx.Roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "需要角色: " + strings.Join(roles, " 或 "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAuthContext 从上下文获取鉴权信息的辅助函数
func GetAuthContext(c *gyarn.Context) (*AuthContext, bool) {
	authCtx, exists := c.Get("auth_context")
	if !exists {
		return nil, false
	}
	return authCtx.(*AuthContext), true
}

// HasPermission 检查当前用户是否有指定权限
func HasPermission(c *gyarn.Context, resource, action string) bool {
	authCtx, exists := GetAuthContext(c)
	if !exists {
		return false
	}

	for _, perm := range authCtx.Permissions {
		if (perm.Resource == resource || perm.Resource == "*") &&
			(perm.Action == action || perm.Action == "*") {
			return true
		}
	}
	return false
}

// HasRole 检查当前用户是否有指定角色
func HasRole(c *gyarn.Context, role string) bool {
	authCtx, exists := GetAuthContext(c)
	if !exists {
		return false
	}

	for _, userRole := range authCtx.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}
