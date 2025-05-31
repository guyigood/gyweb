# GyWeb - 高性能Go Web框架

一个简洁、高效、功能完整的Go Web框架，提供企业级应用开发所需的全套功能。

## ✨ 功能特性

### 🚀 核心功能
- **高性能路由**: 基于前缀树的高效路由匹配
- **中间件支持**: 灵活的中间件机制，支持全局和分组中间件
- **模板引擎**: 内置HTML模板渲染，支持自定义模板函数
- **静态文件服务**: 高效的静态资源服务
- **分组路由**: 支持路由分组，便于API版本管理

### 🛡️ 安全防护
- **跨域处理**: 完整的CORS支持
- **认证授权**: JWT认证中间件
- **请求验证**: 参数验证和数据绑定
- **安全头**: 自动添加安全响应头

### 🔧 实用工具
- **数据库集成**: 支持MySQL、PostgreSQL、SQLite
- **Redis支持**: 完整的Redis操作封装
- **日志系统**: 结构化日志记录
- **配置管理**: 环境变量和配置文件支持
- **优雅关闭**: 服务器优雅关闭处理

### 📱 第三方服务集成
- **微信支付**: 完整的微信支付API集成
- **支付宝**: 支付宝支付接口支持
- **微信小程序**: 小程序登录、用户信息、模板消息
- **微信公众号**: 用户管理、消息推送、菜单管理、二维码生成、网页授权
- **钉钉集成**: 企业应用、消息推送、审批流程
- **Excel服务**: Excel文件导入导出、数据映射、样式配置

### 🌐 WebSocket支持
- **实时通信**: 内置WebSocket支持
- **连接管理**: 自动连接池管理
- **消息广播**: 支持群组消息广播

## 📦 快速开始

### 安装

```bash
go mod init your-project
go get github.com/guyigood/gyweb
```

### 依赖管理

如果使用Excel服务，需要安装额外依赖：

```bash
go get github.com/xuri/excelize/v2
```

### 基础使用

```go
package main

import (
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/gyarn"
    "github.com/guyigood/gyweb/core/middleware"
)

func main() {
    r := engine.New()
    
    // 使用中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // 基础路由
    r.GET("/", func(c *gyarn.Context) {
        c.JSON(200, gyarn.H{
            "message": "Hello GyWeb!",
        })
    })
    
    // 启动服务器
    r.Run(":8080")
}
```

## 🔌 第三方服务集成

### 微信公众号

```go
import "github.com/guyigood/gyweb/core/services/wechat"

config := &wechat.WechatConfig{
    AppID:     "your_app_id",
    AppSecret: "your_app_secret",
    Token:     "your_token",
}

wechatClient := wechat.NewWechat(config)

// 获取用户信息
userInfo, err := wechatClient.GetUserInfo("openid")

// 发送模板消息
msg := &wechat.TemplateMessage{
    ToUser:     "openid",
    TemplateID: "template_id",
    Data:       templateData,
}
wechatClient.SendTemplateMessage(msg)
```

### Excel操作

```go
import "github.com/guyigood/gyweb/core/services/excel"

// 创建Excel服务
excelService := excel.NewExcelService()
defer excelService.Close()

// 导入数据
var users []User
result, err := excelService.ImportData(importOptions, &users)

// 导出数据
err = excelService.ExportData(users, exportOptions)
fileData, err := excelService.GetBytes()
```

### 支付集成

```go
import "github.com/guyigood/gyweb/core/services/payment"

// 微信支付
wechatPay := payment.NewWechatPay(wechatPayConfig)
resp, err := wechatPay.UnifiedOrder(orderReq)

// 支付宝
alipay, err := payment.NewAlipay(alipayConfig)
payURL, err := alipay.TradePagePay(pagePayReq)
```

## 📁 项目结构

```
your-project/
├── main.go                 # 应用入口
├── core/                   # 框架核心
│   ├── engine/            # 引擎
│   ├── gyarn/             # 上下文
│   ├── middleware/        # 中间件
│   └── services/          # 第三方服务
│       ├── payment/       # 支付服务
│       ├── wechat/        # 微信公众号
│       ├── miniprogram/   # 微信小程序
│       ├── dingtalk/      # 钉钉集成
│       └── excel/         # Excel服务
├── examples/              # 示例代码
├── docs/                  # 文档
└── README.md
```

## 🎯 中间件

### 内置中间件

```go
// 日志中间件
r.Use(middleware.Logger())

// 错误恢复
r.Use(middleware.Recovery())

// 跨域处理
r.Use(middleware.CORS())

// JWT认证
r.Use(middleware.JWT("your-secret-key"))

// 限流
r.Use(middleware.RateLimit(100)) // 每分钟100次请求
```

### 自定义中间件

```go
func CustomMiddleware() gyarn.HandlerFunc {
    return func(c *gyarn.Context) {
        // 前置处理
        start := time.Now()
        
        c.Next()
        
        // 后置处理
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    }
}

r.Use(CustomMiddleware())
```

## 🗄️ 数据库操作

```go
import "github.com/guyigood/gyweb/core/database"

// 初始化数据库
db, err := database.NewMySQLDB(config)

// GORM集成
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
}

// 自动迁移
db.AutoMigrate(&User{})

// CRUD操作
var user User
db.First(&user, 1)
db.Create(&User{Name: "John"})
```

## 📊 示例项目

### 第三方服务集成示例

```bash
# 运行支付和小程序示例
go run examples/third_party_services_example.go

# 运行微信公众号和Excel示例  
go run examples/additional_services_example.go
```

### API端点

#### 微信公众号
- `GET /api/wechat/verify` - 验证微信服务器
- `GET /api/wechat/user/:openid` - 获取用户信息
- `POST /api/wechat/message/template` - 发送模板消息
- `POST /api/wechat/qrcode` - 生成二维码
- `POST /api/wechat/menu` - 创建菜单

#### Excel操作
- `POST /api/excel/import` - 导入Excel数据
- `GET /api/excel/export` - 导出Excel数据
- `POST /api/excel/template` - 生成Excel模板

#### 支付服务
- `POST /api/pay/wechat/create` - 创建微信支付订单
- `POST /api/pay/alipay/create` - 创建支付宝订单
- `POST /api/pay/*/notify` - 支付回调处理

## 📚 文档

- [快速开始指南](docs/getting_started.md)
- [第三方服务集成](docs/third_party_services.md)
- [附加服务集成](docs/additional_services.md)
- [中间件开发](docs/middleware.md)
- [数据库操作](docs/database.md)

## 🤝 贡献

欢迎提交问题和拉取请求。对于重大更改，请先开issue讨论您希望进行的更改。

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🌟 特别鸣谢

感谢所有为这个项目做出贡献的开发者们！

---

## 🔗 相关链接

- [GitHub仓库](https://github.com/guyigood/gyweb)
- [问题反馈](https://github.com/guyigood/gyweb/issues)
- [讨论社区](https://github.com/guyigood/gyweb/discussions)

让我们一起构建更好的Go Web应用！ 🚀 