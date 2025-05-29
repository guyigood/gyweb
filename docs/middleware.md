# 中间件说明文档

## 认证中间件

认证中间件提供了统一的认证管理接口，支持多种认证方式，并内置调试功能。

### 快速开始

```go
// 创建认证中间件
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{
        SecretKey: "your-secret-key",
        TokenLookup: "header:Authorization",
    }).
    Build()

// 在路由中使用
r := engine.New()
r.Use(auth)
```

### 认证方式

1. **JWT认证**
```go
jwtAuth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{
        SecretKey: "your-secret-key",
        TokenLookup: "header:Authorization", // header/query/cookie
        TokenHeadName: "Bearer",
    }).
    Build()
```

2. **Session认证**
```go
sessionAuth := middleware.NewAuthManager().
    UseSession().
    Build()
```

3. **Basic认证**
```go
basicAuth := middleware.NewAuthManager().
    UseBasic(map[string]string{
        "admin": "password",
    }).
    Build()
```

4. **自定义认证**
```go
customAuth := middleware.NewAuthManager().
    UseCustom(func(c *context.Context) bool {
        return c.Request.Header.Get("X-Token") != ""
    }).
    Build()
```

### 白名单配置

支持三种匹配方式：精确匹配、前缀匹配、正则匹配。

```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    AddWhitelist(
        []string{"/api/public"},     // 精确匹配
        []string{"/static/"},        // 前缀匹配
        []string{`^/api/v1/.*`},     // 正则匹配
    ).
    Build()
```

### 调试功能

1. **启用调试**
```go
// 方式1：环境变量
export GYWEB_DEBUG=true

// 方式2：代码设置
middleware.SetDebug(true)
```

2. **调试信息**
- 请求路径和方法
- 白名单匹配结果
- 认证函数执行状态
- 未授权处理信息

3. **调试输出示例**
```
[GYWEB-DEBUG] [Auth] 127.0.0.1 - GET /api/users: 开始认证检查
[GYWEB-DEBUG] [Whitelist] Path: /api/users, Type: none, Matched: false
[GYWEB-DEBUG] [AuthFunc] 127.0.0.1 - GET /api/users: true
```

### 安全建议

1. 生产环境必须关闭调试模式
2. JWT密钥定期更换
3. 使用HTTPS传输
4. 合理配置白名单范围
5. 避免使用Basic认证（除非内部系统）
6. Session存储使用安全方案

### 性能优化

1. 白名单优先使用精确匹配
2. 减少正则表达式使用
3. 合理设置认证缓存
4. 避免频繁的认证检查
5. 使用高效的Session存储

### 错误处理

1. **自定义未授权响应**
```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    SetUnauthorizedHandler(func(c *context.Context) {
        c.JSON(http.StatusUnauthorized, map[string]interface{}{
            "code": 401,
            "message": "请先登录",
            "redirect": "/login",
        })
    }).
    Build()
```

2. **错误码规范**
- 401: 未认证
- 403: 无权限
- 400: 认证参数错误
- 500: 认证服务异常