# API 说明文档

## 认证中间件 API

### AuthManager

认证管理器，提供统一的认证接口。

```go
// 创建认证管理器
auth := middleware.NewAuthManager()
```

#### 方法列表

1. **UseJWT**
```go
func (m *AuthManager) UseJWT(config *JWTConfig) *AuthManager
```
配置JWT认证
- `config`: JWT配置参数
  - `SecretKey`: JWT密钥
  - `TokenLookup`: token获取位置 (header/query/cookie)
  - `TokenHeadName`: token前缀 (如 "Bearer")

2. **UseSession**
```go
func (m *AuthManager) UseSession() *AuthManager
```
配置Session认证

3. **UseBasic**
```go
func (m *AuthManager) UseBasic(users map[string]string) *AuthManager
```
配置Basic认证
- `users`: 用户名密码映射表

4. **UseCustom**
```go
func (m *AuthManager) UseCustom(authFunc func(*context.Context) bool) *AuthManager
```
配置自定义认证
- `authFunc`: 自定义认证函数

5. **AddWhitelist**
```go
func (m *AuthManager) AddWhitelist(paths, prefixes, patterns []string) *AuthManager
```
添加白名单配置
- `paths`: 精确匹配路径列表
- `prefixes`: 前缀匹配路径列表
- `patterns`: 正则匹配模式列表

6. **SetUnauthorizedHandler**
```go
func (m *AuthManager) SetUnauthorizedHandler(handler func(*context.Context)) *AuthManager
```
设置未授权处理函数
- `handler`: 自定义未授权处理函数

7. **Build**
```go
func (m *AuthManager) Build() context.HandlerFunc
```
构建认证中间件

### 调试 API

1. **SetDebug**
```go
func SetDebug(enable bool)
```
设置调试模式
- `enable`: 是否启用调试模式

2. **IsDebugEnabled**
```go
func IsDebugEnabled() bool
```
检查调试模式状态

### 配置结构

1. **JWTConfig**
```go
type JWTConfig struct {
    SecretKey     string // JWT密钥
    TokenLookup   string // token获取位置
    TokenHeadName string // token前缀
}
```

2. **AuthConfig**
```go
type AuthConfig struct {
    AuthFunc            func(*context.Context) bool           // 认证函数
    WhitelistPaths      []string                              // 白名单路径
    WhitelistPrefixes   []string                              // 白名单前缀
    WhitelistPatterns   []*regexp.Regexp                      // 白名单正则
    UnauthorizedHandler func(*context.Context)                // 未授权处理
}
```

### 使用示例

1. **基础认证**
```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{
        SecretKey: "your-secret-key",
        TokenLookup: "header:Authorization",
    }).
    Build()
```

2. **带白名单的认证**
```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    AddWhitelist(
        []string{"/api/public"},
        []string{"/static/"},
        []string{`^/api/v1/.*`},
    ).
    Build()
```

3. **自定义未授权处理**
```go
auth := middleware.NewAuthManager().
    UseJWT(&middleware.JWTConfig{...}).
    SetUnauthorizedHandler(func(c *context.Context) {
        c.JSON(401, map[string]interface{}{
            "code": 401,
            "message": "请先登录",
        })
    }).
    Build()
```

### 错误码

| 状态码 | 说明 | 处理建议 |
|--------|------|----------|
| 401 | 未认证 | 引导用户登录 |
| 403 | 无权限 | 检查用户权限 |
| 400 | 参数错误 | 检查请求参数 |
| 500 | 服务异常 | 检查服务状态 |

### 注意事项

1. **安全性**
   - 生产环境必须关闭调试模式
   - JWT密钥需要定期更换
   - 使用HTTPS传输
   - 合理配置白名单

2. **性能**
   - 优先使用精确匹配
   - 减少正则表达式
   - 合理使用缓存
   - 避免频繁认证

3. **调试**
   - 开发环境可启用调试
   - 关注认证日志
   - 及时处理异常
   - 定期检查配置

// GetUser retrieves a user by their ID.
//
// Parameters:
//   - id: The unique identifier of the user.
//
// Returns:
//   - *User: A pointer to the User object if found.
//   - error: An error if the user could not be retrieved.
func (s *UserService) GetUser(id string) (*User, error) {
    // implementation
}