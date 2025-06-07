package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/redis/go-redis/v9"
)

// Permission 权限结构
type Permission struct {
	ID          string   `json:"id"`          // 权限ID
	Name        string   `json:"name"`        // 权限名称
	Resource    string   `json:"resource"`    // 资源标识 (如: user, order, product)
	Action      string   `json:"action"`      // 操作类型 (如: create, read, update, delete)
	Conditions  []string `json:"conditions"`  // 权限条件 (如: owner_only, department_only)
	Description string   `json:"description"` // 权限描述
}

// Role 角色结构
type Role struct {
	ID          string       `json:"id"`           // 角色ID
	Name        string       `json:"name"`         // 角色名称
	Permissions []Permission `json:"permissions"`  // 角色权限列表
	ParentRoles []string     `json:"parent_roles"` // 父角色ID列表(支持角色继承)
	IsActive    bool         `json:"is_active"`    // 是否激活
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`   // 更新时间
}

// User 用户结构
type User struct {
	ID          string            `json:"id"`            // 用户ID
	Username    string            `json:"username"`      // 用户名
	Roles       []string          `json:"roles"`         // 用户角色ID列表
	Permissions []Permission      `json:"permissions"`   // 用户直接权限(覆盖角色权限)
	Attributes  map[string]string `json:"attributes"`    // 用户属性(部门、组织等)
	IsActive    bool              `json:"is_active"`     // 是否激活
	ExpiredAt   *time.Time        `json:"expired_at"`    // 过期时间
	LastLoginAt *time.Time        `json:"last_login_at"` // 最后登录时间
}

// AuthorizationRule 授权规则
type AuthorizationRule struct {
	ID         string            `json:"id"`          // 规则ID
	Resource   string            `json:"resource"`    // 资源模式 (支持通配符: user.*, order.detail)
	Action     string            `json:"action"`      // 操作模式 (支持通配符: read, write, *)
	Method     string            `json:"method"`      // HTTP方法 (GET, POST, PUT, DELETE, *)
	Path       string            `json:"path"`        // URL路径模式 (支持正则: /api/users/\d+)
	Conditions map[string]string `json:"conditions"`  // 条件表达式
	AllowRoles []string          `json:"allow_roles"` // 允许的角色
	DenyRoles  []string          `json:"deny_roles"`  // 拒绝的角色
	Priority   int               `json:"priority"`    // 优先级(数字越大优先级越高)
	IsActive   bool              `json:"is_active"`   // 是否激活
}

// AuthService 鉴权服务
type AuthService struct {
	redis       *redis.Client
	roles       map[string]*Role       // 角色缓存
	rules       []*AuthorizationRule   // 授权规则缓存
	permissions map[string]*Permission // 权限缓存
	users       map[string]*User       // 用户缓存 (生产环境建议使用数据库)
	mutex       sync.RWMutex           // 读写锁
	config      *AuthServiceConfig     // 配置
}

// AuthServiceConfig 鉴权服务配置
type AuthServiceConfig struct {
	RedisClient      *redis.Client `json:"-"`                  // Redis客户端
	CachePrefix      string        `json:"cache_prefix"`       // 缓存前缀
	CacheExpiration  time.Duration `json:"cache_expiration"`   // 缓存过期时间
	EnableSuperAdmin bool          `json:"enable_super_admin"` // 是否启用超级管理员
	SuperAdminRole   string        `json:"super_admin_role"`   // 超级管理员角色
	DefaultDeny      bool          `json:"default_deny"`       // 默认拒绝策略
}

// AuthContext 认证上下文
type AuthContext struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Roles       []string          `json:"roles"`
	Permissions []Permission      `json:"permissions"`
	Attributes  map[string]string `json:"attributes"`
	RequestPath string            `json:"request_path"`
	Method      string            `json:"method"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
}

// NewAuthService 创建鉴权服务
func NewAuthService(config *AuthServiceConfig) *AuthService {
	if config == nil {
		config = &AuthServiceConfig{
			CachePrefix:      "gyweb:auth:",
			CacheExpiration:  30 * time.Minute,
			EnableSuperAdmin: true,
			SuperAdminRole:   "super_admin",
			DefaultDeny:      true,
		}
	}

	service := &AuthService{
		redis:       config.RedisClient,
		roles:       make(map[string]*Role),
		rules:       make([]*AuthorizationRule, 0),
		permissions: make(map[string]*Permission),
		users:       make(map[string]*User),
		config:      config,
	}

	// 初始化默认权限和角色
	service.initializeDefaults()

	return service
}

// initializeDefaults 初始化默认权限和角色
func (s *AuthService) initializeDefaults() {
	// 默认权限
	defaultPermissions := []*Permission{
		{ID: "user.create", Name: "创建用户", Resource: "user", Action: "create"},
		{ID: "user.read", Name: "查看用户", Resource: "user", Action: "read"},
		{ID: "user.update", Name: "更新用户", Resource: "user", Action: "update"},
		{ID: "user.delete", Name: "删除用户", Resource: "user", Action: "delete"},
		{ID: "role.create", Name: "创建角色", Resource: "role", Action: "create"},
		{ID: "role.read", Name: "查看角色", Resource: "role", Action: "read"},
		{ID: "role.update", Name: "更新角色", Resource: "role", Action: "update"},
		{ID: "role.delete", Name: "删除角色", Resource: "role", Action: "delete"},
		{ID: "system.admin", Name: "系统管理", Resource: "system", Action: "*"},
	}

	for _, perm := range defaultPermissions {
		s.permissions[perm.ID] = perm
	}

	// 默认角色
	if s.config.EnableSuperAdmin {
		s.roles[s.config.SuperAdminRole] = &Role{
			ID:          s.config.SuperAdminRole,
			Name:        "超级管理员",
			Permissions: *(*[]Permission)(defaultPermissions),
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// 普通用户角色
	s.roles["user"] = &Role{
		ID:   "user",
		Name: "普通用户",
		Permissions: []Permission{
			*s.permissions["user.read"],
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 默认授权规则
	s.rules = []*AuthorizationRule{
		{
			ID:         "public_access",
			Resource:   "*",
			Action:     "read",
			Method:     "GET",
			Path:       "^/api/public/.*",
			AllowRoles: []string{"*"}, // 所有角色
			Priority:   100,
			IsActive:   true,
		},
		{
			ID:         "admin_full_access",
			Resource:   "*",
			Action:     "*",
			Method:     "*",
			Path:       "/api/admin/.*",
			AllowRoles: []string{s.config.SuperAdminRole},
			Priority:   1000,
			IsActive:   true,
		},
	}
}

// AddUser 添加用户
func (s *AuthService) AddUser(user *User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user.IsActive = true
	if user.Attributes == nil {
		user.Attributes = make(map[string]string)
	}

	s.users[user.ID] = user

	// 缓存到Redis
	if s.redis != nil {
		userData, _ := json.Marshal(user)
		key := s.config.CachePrefix + "user:" + user.ID
		s.redis.Set(context.Background(), key, userData, s.config.CacheExpiration)
	}

	return nil
}

// GetUser 获取用户
func (s *AuthService) GetUser(userID string) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 先从内存查找
	if user, exists := s.users[userID]; exists {
		return user, nil
	}

	// 从Redis查找
	if s.redis != nil {
		key := s.config.CachePrefix + "user:" + userID
		userData, err := s.redis.Get(context.Background(), key).Result()
		if err == nil {
			var user User
			if json.Unmarshal([]byte(userData), &user) == nil {
				s.users[userID] = &user // 缓存到内存
				return &user, nil
			}
		}
	}

	return nil, fmt.Errorf("用户不存在: %s", userID)
}

// AddRole 添加角色
func (s *AuthService) AddRole(role *Role) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	role.IsActive = true
	role.UpdatedAt = time.Now()
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now()
	}

	s.roles[role.ID] = role

	// 缓存到Redis
	if s.redis != nil {
		roleData, _ := json.Marshal(role)
		key := s.config.CachePrefix + "role:" + role.ID
		s.redis.Set(context.Background(), key, roleData, s.config.CacheExpiration)
	}

	return nil
}

// AddRule 添加授权规则
func (s *AuthService) AddRule(rule *AuthorizationRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rule.IsActive = true
	s.rules = append(s.rules, rule)

	// 按优先级排序
	for i := len(s.rules) - 1; i > 0; i-- {
		if s.rules[i].Priority > s.rules[i-1].Priority {
			s.rules[i], s.rules[i-1] = s.rules[i-1], s.rules[i]
		} else {
			break
		}
	}

	// 缓存到Redis
	if s.redis != nil {
		ruleData, _ := json.Marshal(rule)
		key := s.config.CachePrefix + "rule:" + rule.ID
		s.redis.Set(context.Background(), key, ruleData, s.config.CacheExpiration)
	}

	return nil
}

// Authorize 权限验证
func (s *AuthService) Authorize(ctx *AuthContext) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 获取用户信息
	user, err := s.GetUser(ctx.UserID)
	if err != nil {
		return false, fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 检查用户是否激活
	if !user.IsActive {
		return false, fmt.Errorf("用户已禁用")
	}

	// 检查用户是否过期
	if user.ExpiredAt != nil && user.ExpiredAt.Before(time.Now()) {
		return false, fmt.Errorf("用户已过期")
	}

	// 超级管理员直接通过
	if s.config.EnableSuperAdmin {
		for _, roleID := range user.Roles {
			if roleID == s.config.SuperAdminRole {
				return true, nil
			}
		}
	}

	// 检查授权规则
	for _, rule := range s.rules {
		if !rule.IsActive {
			continue
		}

		// 匹配路径
		if rule.Path != "" {
			matched, _ := regexp.MatchString(rule.Path, ctx.RequestPath)
			if !matched {
				continue
			}
		}

		// 匹配HTTP方法
		if rule.Method != "*" && rule.Method != "" && rule.Method != ctx.Method {
			continue
		}

		// 匹配资源
		if rule.Resource != "*" && rule.Resource != "" && !s.matchPattern(rule.Resource, ctx.Resource) {
			continue
		}

		// 匹配操作
		if rule.Action != "*" && rule.Action != "" && !s.matchPattern(rule.Action, ctx.Action) {
			continue
		}

		// 检查条件
		if !s.checkConditions(rule.Conditions, ctx, user) {
			continue
		}

		// 检查拒绝角色
		for _, denyRole := range rule.DenyRoles {
			if s.userHasRole(user, denyRole) {
				return false, fmt.Errorf("角色被拒绝: %s", denyRole)
			}
		}

		// 检查允许角色
		for _, allowRole := range rule.AllowRoles {
			if allowRole == "*" || s.userHasRole(user, allowRole) {
				return true, nil
			}
		}
	}

	// 检查用户直接权限
	if s.hasDirectPermission(user, ctx.Resource, ctx.Action) {
		return true, nil
	}

	// 检查角色权限
	if s.hasRolePermission(user, ctx.Resource, ctx.Action) {
		return true, nil
	}

	// 默认策略
	return !s.config.DefaultDeny, nil
}

// matchPattern 模式匹配
func (s *AuthService) matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	// 支持通配符匹配
	if strings.Contains(pattern, "*") {
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		matched, _ := regexp.MatchString("^"+regexPattern+"$", value)
		return matched
	}

	return pattern == value
}

// checkConditions 检查条件
func (s *AuthService) checkConditions(conditions map[string]string, ctx *AuthContext, user *User) bool {
	for key, expectedValue := range conditions {
		switch key {
		case "owner_only":
			// 检查是否为资源拥有者
			if expectedValue == "true" && ctx.UserID != ctx.Attributes["owner_id"] {
				return false
			}
		case "department":
			// 检查部门
			if user.Attributes["department"] != expectedValue {
				return false
			}
		case "time_range":
			// 检查时间范围 (格式: 09:00-18:00)
			if !s.checkTimeRange(expectedValue) {
				return false
			}
		}
	}
	return true
}

// checkTimeRange 检查时间范围
func (s *AuthService) checkTimeRange(timeRange string) bool {
	// 简单的时间范围检查实现
	now := time.Now()
	currentTime := now.Format("15:04")

	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return true
	}

	return currentTime >= parts[0] && currentTime <= parts[1]
}

// userHasRole 检查用户是否有指定角色
func (s *AuthService) userHasRole(user *User, roleID string) bool {
	for _, userRoleID := range user.Roles {
		if userRoleID == roleID {
			return true
		}

		// 检查角色继承
		if role, exists := s.roles[userRoleID]; exists {
			if s.roleHasParent(role, roleID) {
				return true
			}
		}
	}
	return false
}

// roleHasParent 检查角色继承
func (s *AuthService) roleHasParent(role *Role, parentRoleID string) bool {
	for _, parentID := range role.ParentRoles {
		if parentID == parentRoleID {
			return true
		}

		// 递归检查父角色
		if parentRole, exists := s.roles[parentID]; exists {
			if s.roleHasParent(parentRole, parentRoleID) {
				return true
			}
		}
	}
	return false
}

// hasDirectPermission 检查用户直接权限
func (s *AuthService) hasDirectPermission(user *User, resource, action string) bool {
	for _, perm := range user.Permissions {
		if s.matchPermission(&perm, resource, action) {
			return true
		}
	}
	return false
}

// hasRolePermission 检查角色权限
func (s *AuthService) hasRolePermission(user *User, resource, action string) bool {
	for _, roleID := range user.Roles {
		if role, exists := s.roles[roleID]; exists && role.IsActive {
			for _, perm := range role.Permissions {
				if s.matchPermission(&perm, resource, action) {
					return true
				}
			}

			// 检查父角色权限
			if s.checkParentRolePermissions(role, resource, action) {
				return true
			}
		}
	}
	return false
}

// checkParentRolePermissions 检查父角色权限
func (s *AuthService) checkParentRolePermissions(role *Role, resource, action string) bool {
	for _, parentID := range role.ParentRoles {
		if parentRole, exists := s.roles[parentID]; exists && parentRole.IsActive {
			for _, perm := range parentRole.Permissions {
				if s.matchPermission(&perm, resource, action) {
					return true
				}
			}

			// 递归检查父角色
			if s.checkParentRolePermissions(parentRole, resource, action) {
				return true
			}
		}
	}
	return false
}

// matchPermission 匹配权限
func (s *AuthService) matchPermission(perm *Permission, resource, action string) bool {
	resourceMatch := perm.Resource == "*" || perm.Resource == resource || s.matchPattern(perm.Resource, resource)
	actionMatch := perm.Action == "*" || perm.Action == action || s.matchPattern(perm.Action, action)

	return resourceMatch && actionMatch
}

// CreateMiddleware 创建鉴权中间件
func (s *AuthService) CreateMiddleware() gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		// 从上下文获取用户信息 (应该由认证中间件提前设置)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, map[string]interface{}{
				"code":    401,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		// 构建授权上下文
		authCtx := &AuthContext{
			UserID:      userID.(string),
			Username:    c.GetString("username"),
			RequestPath: c.Request.URL.Path,
			Method:      c.Request.Method,
			Attributes:  make(map[string]string),
		}

		// 解析资源和操作 (从路径中提取)
		authCtx.Resource, authCtx.Action = s.parseResourceAction(authCtx.RequestPath, authCtx.Method)

		// 执行权限验证
		allowed, err := s.Authorize(authCtx)
		if err != nil {
			c.JSON(403, map[string]interface{}{
				"code":    403,
				"message": "权限验证失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(403, map[string]interface{}{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseResourceAction 从请求路径解析资源和操作
func (s *AuthService) parseResourceAction(path, method string) (string, string) {
	// 简单的解析逻辑，可以根据实际需求定制
	parts := strings.Split(strings.Trim(path, "/"), "/")

	var resource, action string

	// 解析资源 (例如: /api/users/123 -> resource=user)
	if len(parts) >= 2 && parts[0] == "api" {
		resource = strings.TrimSuffix(parts[1], "s") // 去掉复数
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

// GetUserPermissions 获取用户所有权限
func (s *AuthService) GetUserPermissions(userID string) ([]Permission, error) {
	user, err := s.GetUser(userID)
	if err != nil {
		return nil, err
	}

	permissions := make([]Permission, 0)
	permissionSet := make(map[string]bool)

	// 添加用户直接权限
	for _, perm := range user.Permissions {
		if !permissionSet[perm.ID] {
			permissions = append(permissions, perm)
			permissionSet[perm.ID] = true
		}
	}

	// 添加角色权限
	for _, roleID := range user.Roles {
		if role, exists := s.roles[roleID]; exists && role.IsActive {
			for _, perm := range role.Permissions {
				if !permissionSet[perm.ID] {
					permissions = append(permissions, perm)
					permissionSet[perm.ID] = true
				}
			}
		}
	}

	return permissions, nil
}

// ClearCache 清空缓存
func (s *AuthService) ClearCache() error {
	if s.redis != nil {
		pattern := s.config.CachePrefix + "*"
		keys, err := s.redis.Keys(context.Background(), pattern).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			return s.redis.Del(context.Background(), keys...).Err()
		}
	}

	return nil
}
