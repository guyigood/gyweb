# 增强版OpenAPI文档生成指南

## 概述

增强版的OpenAPI实现解决了跨文件model引用的问题，能够自动发现项目中所有的结构体定义，并生成详细的参数和返回结果说明，实现类似图片中展示的效果。

## 🚀 最新特性：内联Schema展开

**重要更新**：现在支持将schema直接嵌入到接口的输入输出参数说明中，而不是放在页面下方的独立schemas区域！

### 内联展开效果对比

**之前**：
- 参数显示为 `$ref: "#/components/schemas/User"`
- 需要点击跳转到页面下方查看具体字段
- 文档体验不够直观

**现在**：
- 参数直接展开显示完整的字段结构
- 包含字段名称、类型、说明、示例值、是否必需
- 一目了然，无需跳转，完全符合您的需求！

## 核心特性

1. **跨文件模型自动发现** - 递归扫描项目目录，自动注册所有结构体
2. **详细字段说明** - 支持description、example、required等标签
3. **智能类型推导** - 自动处理嵌入结构体、指针、数组等复杂类型
4. **完整的JSON Schema** - 生成符合OpenAPI 3.0规范的完整schema定义
5. **🔥内联Schema展开** - schema直接嵌入接口参数，无需引用跳转

## 快速开始

### 1. 基本设置

```go
package main

import (
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/openapi"
)

func main() {
    e := engine.New()
    
    // 启用增强版OpenAPI（自动支持内联schema）
    docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
        Title:       "我的API",
        Description: "完整的API文档示例，支持内联schema展开",
        Version:     "1.0.0",
    })
    
    // 自动发现所有模型（推荐方式）
    docs.GenerateFromAnnotations("./")
    
    // 或者手动注册特定目录的模型
    docs.AutoDiscoverModels("./model", "./dto", "./controller")
    
    e.Run(":8080")
}
```

### 2. 定义详细的模型结构

#### model/user.go
```go
package model

import "time"

// User 用户信息
type User struct {
    ID          int64     `json:"id" description:"用户ID" example:"1"`
    Account     string    `json:"account" description:"账号" example:"admin" binding:"required"`
    Password    string    `json:"password" description:"密码" example:"123456" binding:"required"`
    Device      string    `json:"device" description:"设备" example:"web"`
    ValidCode   string    `json:"validCode" description:"验证码" example:"1234"`
    ValidCodeReqNo string `json:"validCodeReqNo" description:"验证码请求号" example:"req123"`
    Name        string    `json:"name" description:"用户名" example:"张三"`
    Email       string    `json:"email" description:"邮箱" example:"user@example.com"`
    Phone       string    `json:"phone" description:"手机号" example:"13800138000"`
    Status      int       `json:"status" description:"状态：0-禁用，1-启用" example:"1"`
    CreatedAt   time.Time `json:"created_at" description:"创建时间"`
    UpdatedAt   time.Time `json:"updated_at" description:"更新时间"`
}

// UserProfile 用户详细信息
type UserProfile struct {
    User                    // 嵌入User结构体
    Avatar      string      `json:"avatar" description:"头像URL" example:"https://example.com/avatar.jpg"`
    Department  string      `json:"department" description:"部门" example:"技术部"`
    Position    string      `json:"position" description:"职位" example:"工程师"`
    Permissions []string    `json:"permissions" description:"权限列表" example:"user:read,user:write"`
}
```

#### dto/auth.go
```go
package dto

// AuthAccountPasswordLoginParam 账号密码登录参数
type AuthAccountPasswordLoginParam struct {
    Account         string `json:"account" description:"账号" example:"admin" binding:"required"`
    Password        string `json:"password" description:"密码" example:"123456" binding:"required"`
    Device          string `json:"device" description:"设备" example:"web"`
    ValidCode       string `json:"validCode" description:"验证码" example:"1234"`
    ValidCodeReqNo  string `json:"validCodeReqNo" description:"验证码请求号" example:"req123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    Token    string `json:"token" description:"访问令牌" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    Expires  int64  `json:"expires" description:"过期时间戳" example:"1640995200"`
    UserInfo User   `json:"userInfo" description:"用户信息"`
}

// StandardResponse 标准响应格式
type StandardResponse struct {
    Code    int         `json:"code" description:"响应码：0-成功，其他-错误" example:"0"`
    Message string      `json:"message" description:"响应消息" example:"操作成功"`
    Data    interface{} `json:"data,omitempty" description:"响应数据"`
}
```

### 3. 使用增强的API注解（支持内联展开）

#### controller/auth.go
```go
package controller

import (
    "github.com/guyigood/gyweb/core/gyarn"
    "your-project/dto"
    "your-project/model"
)

// DoLogin 用户登录
// @Summary B端账号密码登录
// @Description 使用账号密码进行用户认证登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param authAccountPasswordLoginParam body dto.AuthAccountPasswordLoginParam true "登录参数"
// @Success 200 {object} dto.StandardResponse{data=dto.LoginResponse} "登录成功"
// @Failure 400 {object} dto.StandardResponse "参数错误"
// @Failure 401 {object} dto.StandardResponse "认证失败"
// @Failure 500 {object} dto.StandardResponse "服务器错误"
// @Router /auth/b/doLogin [post]
func DoLogin(c *gyarn.Context) {
    var param dto.AuthAccountPasswordLoginParam
    if err := c.ShouldBindJSON(&param); err != nil {
        c.JSON(400, dto.StandardResponse{
            Code:    400,
            Message: "参数错误",
        })
        return
    }
    
    // 业务逻辑...
    
    response := dto.LoginResponse{
        Token:   "generated-jwt-token",
        Expires: 1640995200,
        UserInfo: model.User{
            ID:      1,
            Account: param.Account,
            Name:    "测试用户",
        },
    }
    
    c.JSON(200, dto.StandardResponse{
        Code:    0,
        Message: "登录成功",
        Data:    response,
    })
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1) example(1)
// @Param size query int false "每页数量" default(10) example(10)
// @Param keyword query string false "搜索关键词" example("张三")
// @Success 200 {object} dto.StandardResponse{data=[]model.UserProfile} "获取成功"
// @Failure 401 {object} dto.StandardResponse "未授权"
// @Failure 500 {object} dto.StandardResponse "服务器错误"
// @Security ApiKeyAuth
// @Router /user/list [get]
func GetUserList(c *gyarn.Context) {
    // 实现逻辑...
}
```

## 🎯 内联Schema展开详解

### 支持的注解格式

1. **简单模型展开**
```go
// @Param data body User true "用户信息"
// @Success 200 {object} User "用户信息"
```

2. **嵌套模型展开**
```go
// @Success 200 {object} dto.StandardResponse{data=User} "成功响应"
// @Success 200 {object} dto.StandardResponse{data=[]User} "用户列表响应"
```

3. **数组模型展开**
```go
// @Success 200 {array} User "用户列表"
```

### 展开效果

使用内联schema后，在Swagger UI中您将看到：

**请求参数**：
- 参数名称：authAccountPasswordLoginParam
- 类型：object
- 必需：true
- 说明：登录参数
- **展开的字段结构**：
  - account (string, required): 账号 [example: "admin"]
  - password (string, required): 密码 [example: "123456"]
  - device (string): 设备 [example: "web"]
  - validCode (string): 验证码 [example: "1234"]
  - validCodeReqNo (string): 验证码请求号 [example: "req123"]

**响应结果**：
- 状态码：200
- 说明：登录成功
- **展开的响应结构**：
  - code (integer): 响应码：0-成功，其他-错误 [example: 0]
  - message (string): 响应消息 [example: "操作成功"]
  - data (object): 响应数据
    - token (string): 访问令牌 [example: "eyJhbGci..."]
    - expires (integer): 过期时间戳 [example: 1640995200]
    - userInfo (object): 用户信息
      - id (integer): 用户ID [example: 1]
      - account (string): 账号 [example: "admin"]
      - name (string): 用户名 [example: "张三"]
      - ... (完整展开所有字段)

## 生成效果

使用增强版OpenAPI后，您将获得：

1. **✅ 完整的内联参数文档** - 所有参数直接展开显示，包含字段名称、类型、说明、示例值、是否必需
2. **✅ 详细的内联响应结构** - 响应schema完全展开，支持嵌套对象和数组
3. **✅ 跨文件模型引用** - 自动发现和引用项目中任何位置的结构体
4. **✅ 智能类型处理** - 正确处理数组、指针、嵌入结构体等复杂类型
5. **✅ 零配置体验** - 一行代码启用，自动发现所有模型并内联展开

访问 `http://localhost:8080/swagger` 即可看到完全内联展开的API文档界面，完全符合您的需求！

## 注意事项

1. 确保结构体字段使用正确的json标签
2. 使用description标签添加字段说明
3. 使用example标签提供示例值
4. 使用binding标签标记必需字段
5. 在注释中使用标准的swagger注解格式
6. **新特性**：所有schema现在都会自动内联展开，无需额外配置

这样就能生成像您要求的那样，schema直接嵌入到接口参数说明中的完整API文档了！ 