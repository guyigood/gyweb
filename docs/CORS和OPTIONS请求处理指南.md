# CORS 和 OPTIONS 请求处理指南

## 问题解决

### 问题1: "superfluous response.WriteHeader call" 错误

这个错误是由于 HTTP 响应头被重复写入导致的。我们已经在 `gyarn.Context` 中添加了 `statusWritten` 标志来防止重复调用 `WriteHeader`。

**修复内容:**
- 在 `Context` 结构体中添加了 `statusWritten` 字段
- 修改了 `Status` 方法，只在第一次调用时写入状态码
- 在 `NewContext` 中初始化该字段

### 问题2: OPTIONS 请求未被正确处理

OPTIONS 请求是浏览器在发送跨域请求前的预检请求。我们提供了两种解决方案：

## 使用方法

### 方案1: 使用内置CORS中间件（推荐）

```go
func main() {
    r := engine.New()
    
    // 使用内置CORS中间件 - 已修复OPTIONS处理
    r.Use(middleware.CORS())
    
    // 其他中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    // 注册路由
    r.POST("/api/auth/login", loginHandler)
    
    r.Run(":8080")
}
```

### 方案2: 使用自定义OptionsHandler

```go
package main

import (
    "your-project/lib"
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/middleware"
)

func main() {
    r := engine.New()
    
    // 使用自定义OPTIONS处理器（必须放在第一位）
    r.Use(lib.OptionsHandler())
    
    // 其他中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // 注册路由
    r.POST("/api/auth/login", loginHandler)
    
    r.Run(":8080")
}
```

### 方案3: 中间件顺序优化（当前问题的解决方案）

```go
func main() {
    r := engine.New()
    
    // 重要：CORS必须放在最前面处理OPTIONS请求
    r.Use(middleware.CORS())
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    // 移除lib.OptionsHandler()，因为CORS已经处理了OPTIONS
    // r.Use(lib.OptionsHandler()) // 删除这行
    
    // 认证中间件放在CORS之后
    r.Use(middleware.RateLimit(100))
    CustomAuth(r)
    r.Use(lib.LogDb())
    
    RegRoute(r)
    r.Run(":8080")
}
```

## 中间件执行顺序的重要性

正确的中间件顺序应该是：

1. **CORS/OPTIONS 处理** - 最先处理预检请求
2. **日志记录** - 记录所有请求
3. **错误恢复** - 捕获panic
4. **限流** - 控制请求频率
5. **认证** - 验证用户身份
6. **业务逻辑** - 处理具体请求

## 调试建议

### 启用调试模式

```go
middleware.SetDebug(true)
```

### 检查请求日志

正确配置后，OPTIONS 请求应该显示：
```
[DEBUG] 收到请求: OPTIONS /api/auth/login
[200] OPTIONS /api/auth/login in 1ms
```

而不是：
```
[DEBUG] 收到请求: OPTIONS /api/auth/login
[DEBUG] 未找到路由: OPTIONS /api/auth/login
```

## 常见问题

### Q: 为什么OPTIONS请求还是显示"未找到路由"？

A: 检查中间件顺序，确保CORS中间件在认证中间件之前，并且没有重复使用OPTIONS处理器。

### Q: 前端还是提示CORS错误怎么办？

A: 检查以下几点：
1. CORS中间件是否在最前面
2. 响应头是否正确设置
3. 是否启用了`Access-Control-Allow-Credentials`
4. 前端请求是否包含了正确的头部

### Q: 如何自定义CORS设置？

A: 你可以创建自定义的CORS中间件：

```go
func CustomCORS() middleware.HandlerFunc {
    return func(c *gyarn.Context) {
        origin := c.GetHeader("Origin")
        
        // 允许特定域名
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://yourdomain.com",
        }
        
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                c.SetHeader("Access-Control-Allow-Origin", origin)
                break
            }
        }
        
        c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
        c.SetHeader("Access-Control-Allow-Credentials", "true")
        c.SetHeader("Access-Control-Max-Age", "86400")
        
        if c.Method == "OPTIONS" {
            c.Status(http.StatusOK)
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 最佳实践

1. **简化中间件栈**: 避免重复的功能中间件
2. **正确的顺序**: CORS → 基础中间件 → 认证中间件 → 业务逻辑
3. **使用 Abort()**: 在OPTIONS处理中使用`c.Abort()`停止后续处理
4. **调试输出**: 在开发环境开启调试模式
5. **测试验证**: 使用浏览器开发者工具验证OPTIONS请求的响应 