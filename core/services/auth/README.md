# API 鉴权模块使用说明

## 概述

API鉴权模块提供了完整的认证（Authentication）和授权（Authorization）功能，支持RBAC（基于角色的访问控制）权限管理系统。

## 功能特性

### 🔐 认证功能
- **JWT认证**: 支持JWT Token认证
- **Session认证**: 支持传统Session认证  
- **Basic认证**: 支持HTTP Basic认证
- **自定义认证**: 支持自定义认证逻辑

### 🛡️ 授权功能
- **RBAC权限模型**: 支持用户-角色-权限三层模型
- **角色继承**: 支持角色层级继承
- **动态权限**: 支持运行时权限验证
- **资源级权限**: 支持细粒度的资源访问控制
- **条件权限**: 支持基于条件的权限验证（时间、部门、资源所有者等）

### 📋 权限管理
- **权限缓存**: 支持Redis缓存提升性能
- **白名单机制**: 支持路径白名单配置
- **权限规则**: 支持复杂的授权规则配置
- **实时更新**: 支持权限的实时更新和缓存刷新

## 快速开始

### 1. 基础设置

```go
package main

import (
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/middleware"
    "github.com/guyigood/gyweb/core/services/auth"
    "github.com/redis/go-redis/v9"
    "time"
)

func main() {
    // 创建引擎
    r := engine.New()
    
    // 配置Redis（可选）
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建鉴权服务
    authService := auth.NewAuthService(&auth.AuthServiceConfig{
        RedisClient:      redisClient,
        CachePrefix:      "myapp:auth:",
        CacheExpiration:  30 * time.Minute,
        EnableSuperAdmin: true,
        SuperAdminRole:   "super_admin",
        DefaultDeny:      true,
    })
    
    // 配置JWT认证
    jwtConfig := &middleware.JWTConfig{
        SecretKey:   "your-secret-key",
        TokenLookup: "header:Authorization",
        TokenHeadName: "Bearer",
        ExpiresIn:   24 * time.Hour,
    }
    
    // 创建认证鉴权中间件链
    authChain := auth.CreateAuthChain(jwtConfig, authService)
    
    // 应用到路由组
    api := r.Group("/api")
    api.Use(authChain...)
    
    // 定义路由
    api.GET("/users", getUsers)
    api.POST("/users", auth.RequirePermission("user", "create"), createUser)
    api.DELETE("/users/:id", auth.RequireRole("admin"), deleteUser)
    
    r.Run(":8080")
}
```

### 2. 用户和角色管理

```go
// 添加用户
user := &auth.User{
    ID:       "user001",
    Username: "john_doe",
    Roles:    []string{"user", "editor"},
    Attributes: map[string]string{
        "department": "engineering",
        "level":      "senior",
    },
}
authService.AddUser(user)

// 添加角色
role := &auth.Role{
    ID:   "editor",
    Name: "编辑者", 
    Permissions: []auth.Permission{
        {ID: "article.create", Resource: "article", Action: "create"},
        {ID: "article.read", Resource: "article", Action: "read"},
        {ID: "article.update", Resource: "article", Action: "update"},
    },
    ParentRoles: []string{"user"}, // 继承user角色
}
authService.AddRole(role)
```

### 3. 权限规则配置

```go
// 添加授权规则
rule := &auth.AuthorizationRule{
    ID:       "admin_full_access",
    Resource: "*",
    Action:   "*", 
    Method:   "*",
    Path:     "/api/admin/.*",
    AllowRoles: []string{"admin", "super_admin"},
    Priority: 1000,
    IsActive: true,
}
authService.AddRule(rule)

// 条件权限规则
ownerRule := &auth.AuthorizationRule{
    ID:       "owner_only_access",
    Resource: "document",
    Action:   "update",
    Method:   "PUT",
    Path:     "/api/documents/\\d+",
    Conditions: map[string]string{
        "owner_only": "true",
        "time_range": "09:00-18:00",
    },
    AllowRoles: []string{"user"},
    Priority:   500,
    IsActive:   true,
}
authService.AddRule(ownerRule)
```

## 详细使用指南

### 认证配置

#### JWT认证配置
```go
jwtConfig := &middleware.JWTConfig{
    SecretKey:     "your-secret-key-at-least-32-chars",
    TokenLookup:   "header:Authorization", // 或 "query:token" 或 "cookie:auth_token"
    TokenHeadName: "Bearer",
    ExpiresIn:     24 * time.Hour,
}

// 生成JWT Token
token, err := middleware.GenerateJWT(jwtConfig, "user123", "john_doe", "admin")
```

#### Session认证配置
```go
sessionConfig := &middleware.SessionConfig{
    SecretKey: "your-session-secret",
    MaxAge:    3600, // 1小时
    Path:      "/",
    Secure:    false, // 生产环境设为true
    HttpOnly:  true,
}

authChain := auth.CreateSessionAuthChain(sessionConfig, authService)
```

### 权限装饰器

#### 权限要求装饰器
```go
// 要求特定权限
r.POST("/api/articles", 
    auth.RequirePermission("article", "create"),
    createArticle)

// 要求特定角色
r.DELETE("/api/users/:id",
    auth.RequireRole("admin", "super_admin"), 
    deleteUser)

// 组合使用
r.PUT("/api/articles/:id",
    auth.RequireRole("editor"),
    auth.RequirePermission("article", "update"),
    updateArticle)
```

#### 在处理函数中检查权限
```go
func getArticles(c *gyarn.Context) {
    // 获取鉴权上下文
    authCtx, exists := auth.GetAuthContext(c)
    if !exists {
        c.Error(401, "未认证")
        return
    }
    
    // 检查权限
    if !auth.HasPermission(c, "article", "read") {
        c.Error(403, "权限不足")
        return 
    }
    
    // 检查角色
    if auth.HasRole(c, "admin") {
        // 管理员可以看到所有文章
        articles := getAllArticles()
        c.Success(articles)
    } else {
        // 普通用户只能看到自己的文章
        articles := getUserArticles(authCtx.UserID)
        c.Success(articles)
    }
}
```

### 白名单配置

```go
// 创建简化中间件并配置白名单
authMiddleware := auth.NewSimpleAuthMiddleware(authService).
    SkipPaths("/health", "/metrics", "/api/login", "/api/register").
    SkipPrefix("/static/", "/public/", "/docs/")

r.Use(authMiddleware.Handler())
```

### 动态权限管理

```go
// 运行时添加权限
permission := &auth.Permission{
    ID:          "report.export",
    Name:        "导出报表",
    Resource:    "report", 
    Action:      "export",
    Description: "允许导出报表数据",
}

// 给角色添加权限
role, _ := authService.GetRole("manager")
role.Permissions = append(role.Permissions, *permission)
authService.AddRole(role)

// 清空缓存使权限立即生效
authService.ClearCache()
```

### 条件权限示例

```go
// 时间限制权限
timeRule := &auth.AuthorizationRule{
    ID:       "business_hours_only",
    Resource: "financial",
    Action:   "*",
    Method:   "*", 
    Path:     "/api/financial/.*",
    Conditions: map[string]string{
        "time_range": "09:00-18:00",
    },
    AllowRoles: []string{"accountant"},
    Priority:   800,
    IsActive:   true,
}

// 部门限制权限 
deptRule := &auth.AuthorizationRule{
    ID:       "dept_data_access",
    Resource: "employee",
    Action:   "read",
    Conditions: map[string]string{
        "department": "hr",
    },
    AllowRoles: []string{"hr_manager"},
    Priority:   600,
    IsActive:   true,
}

// 资源所有者权限
ownerRule := &auth.AuthorizationRule{
    ID:       "owner_document_access", 
    Resource: "document",
    Action:   "update",
    Conditions: map[string]string{
        "owner_only": "true",
    },
    AllowRoles: []string{"user"},
    Priority:   400,
    IsActive:   true,
}

authService.AddRule(timeRule)
authService.AddRule(deptRule) 
authService.AddRule(ownerRule)
```

## 高级特性

### 1. 角色继承

```go
// 定义角色层次
adminRole := &auth.Role{
    ID:   "admin",
    Name: "管理员",
    Permissions: []auth.Permission{
        {Resource: "system", Action: "*"},
    },
}

superAdminRole := &auth.Role{
    ID:          "super_admin", 
    Name:        "超级管理员",
    ParentRoles: []string{"admin"}, // 继承admin角色的所有权限
    Permissions: []auth.Permission{
        {Resource: "*", Action: "*"}, // 额外的权限
    },
}
```

### 2. 自定义资源解析器

```go
customParser := func(path, method string) (string, string) {
    // 自定义解析逻辑
    if strings.HasPrefix(path, "/api/v1/") {
        parts := strings.Split(path, "/")
        if len(parts) >= 4 {
            return parts[3], methodToAction(method) 
        }
    }
    return "unknown", "unknown"
}

authMiddleware := auth.NewSimpleAuthMiddleware(authService)
// 注意：当前简化版本不支持自定义解析器，需要在service中设置
```

### 3. 性能优化

```go
// 使用Redis缓存
config := &auth.AuthServiceConfig{
    RedisClient:     redisClient,
    CachePrefix:     "myapp:auth:",
    CacheExpiration: 30 * time.Minute,
}

// 批量权限检查
permissions, err := authService.GetUserPermissions(userID)
if err != nil {
    // 处理错误
}

// 预加载用户权限到上下文
// 这样可以避免每次请求都查询数据库
```

## 安全建议

### 1. 密钥管理
```go
// 使用环境变量存储密钥
secretKey := os.Getenv("JWT_SECRET_KEY")
if secretKey == "" {
    log.Fatal("JWT_SECRET_KEY environment variable is required")
}

// 确保密钥足够复杂（至少32字符）
if len(secretKey) < 32 {
    log.Fatal("JWT secret key must be at least 32 characters")
}
```

### 2. HTTPS配置
```go
// 生产环境强制使用HTTPS
if gin.Mode() == gin.ReleaseMode {
    sessionConfig.Secure = true
    sessionConfig.HttpOnly = true
}
```

### 3. 权限最小化原则
```go
// 默认拒绝策略
config.DefaultDeny = true

// 定期审查权限
func auditPermissions() {
    users := getAllUsers()
    for _, user := range users {
        permissions, _ := authService.GetUserPermissions(user.ID)
        logUserPermissions(user.ID, permissions)
    }
}
```

## 错误处理

```go
// 自定义错误处理
func customErrorHandler(c *gyarn.Context, err error, code int) {
    logAuthError(c.Request.RemoteAddr, c.Request.URL.Path, err)
    
    switch code {
    case 401:
        c.JSON(http.StatusUnauthorized, map[string]interface{}{
            "error": "authentication_required",
            "message": "请登录后访问",
            "redirect": "/login",
        })
    case 403:
        c.JSON(http.StatusForbidden, map[string]interface{}{
            "error": "insufficient_permissions", 
            "message": "权限不足，请联系管理员",
        })
    }
}
```

## 监控和日志

```go
// 权限验证日志
func logAuthEvent(userID, resource, action string, allowed bool) {
    log.Printf("AUTH: user=%s resource=%s action=%s allowed=%v", 
        userID, resource, action, allowed)
}

// 集成到鉴权服务中
// 在Authorize方法中添加日志记录
```

## 最佳实践

1. **分层权限设计**: 使用角色继承减少权限配置复杂度
2. **缓存策略**: 合理使用Redis缓存提升性能  
3. **权限粒度**: 根据业务需求设计合适的权限粒度
4. **白名单优先**: 对公开资源使用白名单而不是权限控制
5. **定期审计**: 定期审查用户权限，及时清理无用权限
6. **错误处理**: 提供友好的错误提示和处理流程
7. **监控告警**: 监控异常的权限访问行为

## 性能参考

- **内存缓存**: 权限验证响应时间 < 1ms
- **Redis缓存**: 权限验证响应时间 < 5ms  
- **数据库查询**: 权限验证响应时间 < 50ms
- **并发支持**: 支持万级并发权限验证

建议在高并发场景下使用Redis缓存，并合理设置缓存过期时间。 