# GyWeb OpenAPI 注解使用指南

## 📋 目录

- [简介](#简介)
- [基本注解](#基本注解)
- [参数注解](#参数注解)
- [请求体注解](#请求体注解)
- [响应注解](#响应注解)
- [安全注解](#安全注解)
- [结构体标签](#结构体标签)
- [完整示例](#完整示例)
- [跨文件结构体处理](#跨文件结构体处理)
- [最佳实践](#最佳实践)

## 🚀 简介

GyWeb OpenAPI 支持通过注释自动生成 API 文档。只需要在函数上方添加特定格式的注释，就能自动生成完整的 Swagger/OpenAPI 文档。

### 启用 OpenAPI

```go
// 启用OpenAPI - 一行代码！
docs := openapi.EnableOpenAPI(r, openapi.OpenAPIConfig{
    Title:   "我的API",
    Version: "1.0.0",
})

// 从注解生成文档 - 关键的一行！
err := docs.GenerateFromAnnotations("./")
if err != nil {
    fmt.Println("OpenApi 引擎生成失败！", err)
    return
}
```

## 📝 基本注解

### @Summary - API摘要
简短描述API的功能

```go
// @Summary 获取用户列表
```

### @Description - API详细描述
详细描述API的功能和用途

```go
// @Description 获取系统中所有用户的分页列表，支持搜索和排序
```

### @Tags - API标签
用于对API进行分组，多个标签用逗号分隔

```go
// @Tags 用户管理
// @Tags 用户管理, 系统管理
```

### @Router - 路由信息
定义API的路径和HTTP方法

```go
// @Router /api/users [get]
// @Router /api/users/{id} [put]
// @Router /api/users [post]
```

### @Deprecated - 标记已弃用
标记API为已弃用状态

```go
// @Deprecated
```

## 🔧 参数注解

### @Param - 参数定义

**格式**：`@Param name in type required "description" default(value) example(value)`

#### 参数位置（in）：
- `query` - 查询参数
- `path` - 路径参数
- `header` - 头部参数
- `cookie` - Cookie参数
- `body` - 请求体参数

#### 参数类型（type）：
- `string` - 字符串
- `integer` - 整数
- `number` - 浮点数
- `boolean` - 布尔值
- `array` - 数组
- `object` - 对象

#### 示例：

```go
// 查询参数
// @Param page query integer false "页码" default(1) example(1)
// @Param size query integer false "每页数量" default(10) example(10)
// @Param search query string false "搜索关键词" example("张三")

// 路径参数
// @Param id path integer true "用户ID" example(1)

// 头部参数
// @Param Authorization header string true "认证token" example("Bearer eyJhbGci...")
// @Param Content-Type header string true "内容类型" example("application/json")

// 请求体参数
// @Param body body UserCreateRequest true "用户创建请求"
```

## 📤 请求体注解

### @Accept - 接受的内容类型
指定API接受的请求内容类型

```go
// @Accept json
// @Accept multipart/form-data
// @Accept application/xml
```

### 请求体定义

通过 `@Param body` 定义请求体：

```go
// @Param body body UserCreateRequest true "用户创建请求"
```

## 📥 响应注解

### @Produce - 响应内容类型
指定API响应的内容类型

```go
// @Produce json
// @Produce application/xml
```

### @Success - 成功响应

**格式**：`@Success code {type} model "description"`

```go
// @Success 200 {object} User "获取用户成功"
// @Success 200 {array} User "获取用户列表成功"
// @Success 200 {string} string "操作成功"
// @Success 201 {object} UserCreateResponse "用户创建成功"
```

### @Failure - 失败响应

**格式**：`@Failure code {type} model "description"`

```go
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
```

## 🔐 安全注解

### @Security - 安全要求

```go
// @Security BearerAuth
// @Security ApiKeyAuth
// @Security BasicAuth
```

需要先在OpenAPI配置中定义安全方案：

```go
// JWT Bearer Token
docs.GetOpenAPI().AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer Token认证",
})
```

## 🏷️ 结构体标签

在Go结构体中使用标签来增强Schema生成：

```go
type User struct {
    ID        int       `json:"id" description:"用户ID" example:"1"`
    Name      string    `json:"name" description:"用户名" example:"张三"`
    Email     string    `json:"email" description:"邮箱地址" example:"user@example.com"`
    Age       int       `json:"age" description:"年龄" example:"25"`
    IsActive  bool      `json:"is_active" description:"是否激活" example:"true"`
    Tags      []string  `json:"tags" description:"用户标签" example:"[\"developer\",\"admin\"]"`
    Profile   Profile   `json:"profile" description:"用户档案"`
    CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

type Profile struct {
    Avatar   string `json:"avatar" description:"头像URL" example:"https://example.com/avatar.jpg"`
    Bio      string `json:"bio" description:"个人简介" example:"Go语言开发工程师"`
    Location string `json:"location" description:"所在地" example:"北京"`
}
```

### 支持的标签：
- `json` - JSON字段名
- `description` - 字段描述
- `example` - 示例值

## 📋 完整示例

### GET 请求示例

```go
// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取系统中所有用户的分页列表，支持搜索和排序
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query integer false "页码" default(1) example(1)
// @Param size query integer false "每页数量" default(10) example(10)
// @Param search query string false "搜索关键词" example("张三")
// @Param sort query string false "排序字段" example("id")
// @Param order query string false "排序方向" example("asc")
// @Success 200 {object} UserListResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/users [get]
func GetUsers(c *gyarn.Context) {
    // 实现代码...
}
```

### POST 请求示例

```go
// CreateUser 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param body body UserCreateRequest true "用户创建请求"
// @Success 201 {object} User "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 409 {object} ErrorResponse "用户已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/users [post]
func CreateUser(c *gyarn.Context) {
    var req UserCreateRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "请求参数错误")
        return
    }
    
    // 创建用户逻辑...
    user := User{
        Name:  req.Name,
        Email: req.Email,
        Age:   req.Age,
    }
    
    c.Success(user)
}
```

### PUT 请求示例

```go
// UpdateUser 更新用户
// @Summary 更新用户信息
// @Description 根据用户ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path integer true "用户ID" example(1)
// @Param body body UserUpdateRequest true "用户更新请求"
// @Success 200 {object} User "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/{id} [put]
func UpdateUser(c *gyarn.Context) {
    // 实现代码...
}
```

### DELETE 请求示例

```go
// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path integer true "用户ID" example(1)
// @Success 200 {object} SuccessResponse "删除成功"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/{id} [delete]
func DeleteUser(c *gyarn.Context) {
    // 实现代码...
}
```

## 🔗 跨文件结构体处理

### 文件结构示例

```
project/
├── controller/
│   ├── usercontroller.go
│   └── productcontroller.go
├── model/
│   ├── user.go
│   ├── product.go
│   └── request.go
└── types/
    └── response.go
```

### 在注解中引用其他包的结构体

```go
// controller/usercontroller.go
package usercontroller

import (
    "your-project/model"
    "your-project/types"
)

// CreateUser 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param body body model.UserCreateRequest true "用户创建请求"
// @Success 201 {object} model.User "创建成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Router /api/users [post]
func CreateUser(c *gyarn.Context) {
    var req model.UserCreateRequest
    // 实现代码...
}
```

### 结构体定义示例

```go
// model/user.go
package model

type User struct {
    ID       int    `json:"id" description:"用户ID" example:"1"`
    Name     string `json:"name" description:"用户名" example:"张三"`
    Email    string `json:"email" description:"邮箱" example:"user@example.com"`
    Age      int    `json:"age" description:"年龄" example:"25"`
    IsActive bool   `json:"is_active" description:"是否激活" example:"true"`
}

type UserCreateRequest struct {
    Name  string `json:"name" description:"用户名" example:"张三"`
    Email string `json:"email" description:"邮箱" example:"user@example.com"`
    Age   int    `json:"age" description:"年龄" example:"25"`
}

type UserUpdateRequest struct {
    Name     *string `json:"name,omitempty" description:"用户名" example:"李四"`
    Email    *string `json:"email,omitempty" description:"邮箱" example:"lisi@example.com"`
    Age      *int    `json:"age,omitempty" description:"年龄" example:"30"`
    IsActive *bool   `json:"is_active,omitempty" description:"是否激活" example:"false"`
}
```

```go
// types/response.go
package types

type ErrorResponse struct {
    Code int    `json:"code" description:"错误码" example:"400"`
    Msg  string `json:"msg" description:"错误信息" example:"请求参数错误"`
    Data string `json:"data" description:"错误数据" example:"null"`
}

type SuccessResponse struct {
    Code    int         `json:"code" description:"响应码" example:"200"`
    Message string      `json:"message" description:"响应信息" example:"操作成功"`
    Data    interface{} `json:"data" description:"响应数据"`
}

type UserListResponse struct {
    Code int `json:"code" description:"响应码" example:"200"`
    Msg  string `json:"msg" description:"响应信息" example:"success"`
    Data struct {
        Total int          `json:"total" description:"总数" example:"100"`
        List  []model.User `json:"list" description:"用户列表"`
    } `json:"data" description:"响应数据"`
}
```

## 🎯 最佳实践

### 1. 统一的响应结构

定义统一的API响应格式：

```go
// 成功响应
type ApiResponse struct {
    Code    int         `json:"code" description:"响应码" example:"200"`
    Message string      `json:"message" description:"响应信息" example:"success"`
    Data    interface{} `json:"data" description:"响应数据"`
}

// 错误响应
type ErrorResponse struct {
    Code int    `json:"code" description:"错误码" example:"400"`
    Msg  string `json:"msg" description:"错误信息" example:"参数错误"`
    Data string `json:"data" description:"错误数据" example:"null"`
}

// 分页响应
type PageResponse struct {
    Code int `json:"code" description:"响应码" example:"200"`
    Msg  string `json:"msg" description:"响应信息" example:"success"`
    Data struct {
        Total    int         `json:"total" description:"总数"`
        Page     int         `json:"page" description:"当前页"`
        PageSize int         `json:"page_size" description:"每页数量"`
        List     interface{} `json:"list" description:"数据列表"`
    } `json:"data" description:"分页数据"`
}
```

### 2. 标准化的错误码

```go
const (
    // 成功
    CodeSuccess = 200
    
    // 客户端错误
    CodeBadRequest    = 400
    CodeUnauthorized  = 401
    CodeForbidden     = 403
    CodeNotFound      = 404
    
    // 服务端错误
    CodeInternalError = 500
    
    // 业务错误
    CodeParamError    = 1001
    CodeUserNotFound  = 1002
    CodeUserExists    = 1003
)
```

### 3. 完整的文档模板

```go
// FunctionName 函数功能描述
// @Summary API简要描述
// @Description API详细描述，可以多行
// @Tags API分组标签
// @Accept json
// @Produce json
// @Param name query string false "参数描述" default(默认值) example(示例值)
// @Param body body RequestStruct true "请求体描述"
// @Success 200 {object} ResponseStruct "成功描述"
// @Failure 400 {object} ErrorResponse "失败描述"
// @Security BearerAuth
// @Router /api/path [method]
func FunctionName(c *gyarn.Context) {
    // 实现代码...
}
```

### 4. 注释规范

- 注释必须紧邻函数定义，中间不能有空行
- 每个注解独占一行
- 注解顺序建议：基本信息 → 请求参数 → 响应 → 安全 → 路由
- 描述要准确、简洁、有意义
- 示例值要真实、有代表性

### 5. 结构体设计规范

- 使用有意义的字段名
- 添加完整的 JSON 标签
- 提供描述和示例
- 合理使用指针类型处理可选字段
- 遵循 Go 命名约定

## 📚 常用响应类型参考

```go
// 字符串响应
// @Success 200 {string} string "操作成功"

// 对象响应
// @Success 200 {object} User "用户信息"

// 数组响应
// @Success 200 {array} User "用户列表"

// 分页响应
// @Success 200 {object} PageResponse "分页数据"

// 原始类型响应
// @Success 200 {integer} integer "数量"
// @Success 200 {boolean} boolean "是否成功"
// @Success 200 {number} number "金额"
```

## 🔧 常见问题解决

### 1. 结构体无法识别
确保导入了正确的包，并使用完整的包名.结构体名格式

### 2. 文档生成失败
检查注解格式是否正确，特别是 @Router 的格式

### 3. 中文乱码
确保文件编码为 UTF-8

### 4. 嵌套结构体显示不全
确保所有相关结构体都有完整的标签定义

通过遵循以上指南，你就能创建出完整、准确、易读的 API 文档！ 