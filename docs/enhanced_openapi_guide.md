# 增强版OpenAPI文档生成指南

## 概述

增强版的OpenAPI实现解决了跨文件model引用的问题，能够自动发现项目中所有的结构体定义，并生成详细的参数和返回结果说明，实现类似图片中展示的效果。

## 核心特性

1. **跨文件模型自动发现** - 递归扫描项目目录，自动注册所有结构体
2. **详细字段说明** - 支持description、example、required等标签
3. **智能类型推导** - 自动处理嵌入结构体、指针、数组等复杂类型
4. **完整的JSON Schema** - 生成符合OpenAPI 3.0规范的完整schema定义

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
    
    // 启用增强版OpenAPI
    docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
        Title:       "我的API",
        Description: "完整的API文档示例",
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

### 3. 添加详细的API注解

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

### 4. 路由注册

```go
package main

import (
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/openapi"
    "your-project/controller"
)

func main() {
    e := engine.New()
    
    // 启用OpenAPI
    docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
        Title:       "企业管理系统API",
        Description: "完整的企业管理系统API文档",
        Version:     "1.0.0",
    })
    
    // 自动发现所有模型
    docs.GenerateFromAnnotations("./")
    
    // 注册路由
    e.POST("/auth/b/doLogin", controller.DoLogin)
    e.GET("/user/list", controller.GetUserList)
    
    log.Println("API文档地址: http://localhost:8080/swagger")
    e.Run(":8080")
}
```

## 高级用法

### 1. 手动注册复杂模型

```go
// 如果自动发现无法满足需求，可以手动注册
docs.RegisterModels(map[string]interface{}{
    "CustomResponse": CustomResponse{},
    "ComplexData":    ComplexData{},
})
```

### 2. 添加安全认证

```go
docs.AddSecurityScheme("ApiKeyAuth", openapi.SecurityScheme{
    Type: "apiKey",
    In:   "header",
    Name: "Authorization",
})
```

### 3. 自定义响应模板

```go
docs.AddCommonResponses() // 添加常用响应模板
```

## 生成效果

使用增强版OpenAPI后，您将获得：

1. **完整的请求参数文档** - 包含参数名称、类型、说明、示例值、是否必需
2. **详细的响应结构** - 完整的JSON Schema定义，支持嵌套对象
3. **跨文件模型引用** - 自动发现和引用项目中任何位置的结构体
4. **智能类型处理** - 正确处理数组、指针、嵌入结构体等复杂类型
5. **交互式文档** - Swagger UI界面，支持在线测试API

访问 `http://localhost:8080/swagger` 即可查看生成的完整API文档。

## 注意事项

1. 确保结构体字段使用正确的json标签
2. 使用description标签添加字段说明
3. 使用example标签提供示例值
4. 使用binding标签标记必需字段
5. 在注释中使用标准的swagger注解格式

这样就能生成像您图片中展示的那样详细完整的API文档了！ 