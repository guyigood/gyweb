# 中间件使用说明

## 认证中间件 (Auth Middleware)

认证中间件提供了灵活的认证机制，支持多种认证方式和白名单配置。

### 1. 基本使用

```go
// 创建认证管理器
auth := middleware.NewAuthManager()

// 构建认证中间件
authMiddleware := auth.UseJWT(&middleware.JWTConfig{
    SecretKey: "your-secret-key",
    TokenLookup: "header:Authorization",
    TokenHeadName: "Bearer",
}).Build()

// 在路由中使用
r := engine.New()
r.Use(authMiddleware)
```

### 2. 认证方式

#### 2.1 JWT认证
```go
jwtAuth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{
        SecretKey: "your-secret-key",
        TokenLookup: "header:Authorization", // 支持: header, query, cookie
        TokenHeadName: "Bearer",
    }).
    Build()
```

#### 2.2 Session认证
```go
sessionAuth := middleware.NewAuthManager().
    UseSession().
    Build()
```

#### 2.3 Basic认证
```go
basicAuth := middleware.NewAuthManager().
    UseBasic(map[string]string{
        "admin": "password",
        "user": "123456",
    }).
    Build()
```

#### 2.4 自定义认证
```go
customAuth := middleware.NewAuthManager().
    UseCustom(func(c *context.Context) bool {
        // 自定义认证逻辑
        token := c.Request.Header.Get("X-Custom-Token")
        return token != ""
    }).
    Build()
```

### 3. 白名单配置

支持三种白名单配置方式：

```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    AddWhitelist(
        []string{"/public", "/login"},           // 精确匹配路径
        []string{"/static/", "/public/"},        // 前缀匹配
        []string{`^/api/v1/public/.*`},          // 正则匹配
    ).
    Build()
```

### 4. 自定义未授权处理

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

### 5. 调试模式

框架提供了调试模式开关，可以通过环境变量或代码设置：

```go
// 通过环境变量设置
export GYWEB_DEBUG=true

// 或通过代码设置
middleware.SetDebug(true)

// 在认证中间件中会输出详细的认证过程信息
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    Build()
```

调试信息包括：
- 认证请求的路径
- 白名单匹配结果
- 认证函数执行结果
- 未授权处理信息

### 6. 最佳实践

1. 生产环境建议关闭调试模式
2. 合理配置白名单，避免过多正则匹配
3. 根据业务需求选择合适的认证方式
4. 自定义未授权处理时注意返回合适的状态码和信息
5. 定期更新密钥和密码
6. 使用HTTPS保护认证信息

### 7. 注意事项

1. JWT认证需要妥善保管密钥
2. Session认证需要配置合适的存储方式
3. Basic认证仅建议在内部系统使用
4. 自定义认证逻辑需要考虑性能影响
5. 白名单配置要避免过于宽松
6. 生产环境必须关闭调试模式 

```go
// 1. 通过环境变量启用调试模式
export GYWEB_DEBUG=true

// 2. 或通过代码启用调试模式
middleware.SetDebug(true)

// 3. 创建认证中间件
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{
        SecretKey: "your-secret-key",
        TokenLookup: "header:Authorization",
        TokenHeadName: "Bearer",
    }).
    AddWhitelist(
        []string{"/public", "/login"},
        []string{"/static/"},
        []string{`^/api/v1/public/.*`},
    ).
    Build()

// 4. 在路由中使用
r := engine.New()
r.Use(auth)
```