# GyWeb

GyWeb 是一个高性能、轻量级的 Go Web 框架，遵循 Go 语言"简单胜于复杂"的哲学，提供常用 Web 开发功能而不臃肿。

## 特性

- 🚀 高性能：基于 Radix 树的路由，优化的中间件链
- 🎯 轻量级：核心功能精简，易于理解和扩展
- 🔌 中间件：支持中间件链式调用，内置常用中间件
- 📦 模块化：核心功能模块化，易于扩展
- 🛡️ 安全：内置安全防护，防止常见 Web 攻击
- 📝 文档：详细的文档和示例

## 快速开始

### 安装

```bash
go get github.com/yourusername/gyweb
```

### 示例代码

```go
package main

import (
    "net/http"
    "github.com/yourusername/gyweb"
)

func main() {
    // 创建引擎实例
    r := gyweb.New()
    
    // 使用中间件
    r.Use(gyweb.Logger())
    r.Use(gyweb.Recovery())
    
    // 注册路由
    r.GET("/", func(c *gyweb.Context) {
        c.JSON(http.StatusOK, gyweb.H{
            "message": "Welcome to GyWeb!",
        })
    })
    
    // 路由组
    api := r.Group("/api")
    {
        api.GET("/users", func(c *gyweb.Context) {
            c.JSON(http.StatusOK, gyweb.H{
                "users": []string{"Alice", "Bob"},
            })
        })
    }
    
    // 启动服务器
    r.Run(":8080")
}
```

## 核心功能

### 路由系统
- 支持静态路由和动态路由
- 路由分组
- 路由参数解析
- 支持所有 HTTP 方法

### 中间件
- 日志中间件
- 恢复中间件
- CORS 中间件
- 自定义中间件支持

### 上下文
- 请求参数解析
- JSON/HTML/Text 响应
- 文件上传
- Cookie 处理

### 模板渲染
- 支持 HTML 模板
- 模板继承
- 模板变量渲染

## 文档

- [API 文档](docs/api.md)
- [中间件文档](docs/middleware.md)
- [最佳实践](docs/best-practices.md)
- [示例代码](examples/)

## 性能

GyWeb 框架经过精心优化，具有出色的性能表现：

- 路由匹配：使用 Radix 树实现高效路由匹配
- 内存管理：使用对象池减少内存分配
- 并发处理：优化的并发安全设计

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 提交问题和建议
- 提交 Pull Request
- 改进文档
- 分享使用经验

请查看 [贡献指南](CONTRIBUTING.md) 了解详情。

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 作者

- 辜翌 (@guyi)

## 致谢

感谢所有为这个项目做出贡献的开发者！

## 相关项目

- [Gin](https://github.com/gin-gonic/gin) - 参考了部分设计理念
- [Echo](https://github.com/labstack/echo) - 参考了部分设计理念 