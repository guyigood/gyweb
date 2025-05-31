# GyWeb OpenAPI 文档支持

GyWeb框架提供了完整的OpenAPI 3.0.3文档生成和展示功能，支持自动生成API文档、Swagger UI集成、注解解析等特性。

## 功能特性

- ✅ **OpenAPI 3.0.3规范支持** - 完全兼容OpenAPI 3.0.3标准
- ✅ **Swagger UI集成** - 内置美观的API文档界面
- ✅ **注解自动解析** - 支持从Go注释自动生成文档
- ✅ **结构体Schema生成** - 自动从Go结构体生成OpenAPI Schema
- ✅ **多种认证方案** - 支持JWT、API Key、Basic Auth等
- ✅ **流式API构建** - 提供链式调用的文档构建器
- ✅ **常用模板** - 内置常用响应和安全方案模板

## 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/guyigood/gyweb/core/engine"
    "github.com/guyigood/gyweb/core/openapi"
)

func main() {
    // 创建引擎
    e := engine.New()
    
    // 启用OpenAPI支持
    docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
        Title:       "我的API",
        Description: "这是一个示例API",
        Version:     "1.0.0",
        DocsPath:    "/swagger", // 文档访问路径
    })
    
    // 添加服务器信息
    docs.AddServer(openapi.Server{
        URL:         "http://localhost:8080",
        Description: "开发环境",
    })
    
    // 添加标签
    docs.AddTag(openapi.Tag{
        Name:        "用户管理",
        Description: "用户相关的API接口",
    })
    
    // 定义路由和文档
    e.GET("/users", getUsersHandler)
    docs.AddRoute("GET", "/users", openapi.NewDocBuilder().
        Summary("获取用户列表").
        Description("获取所有用户的列表信息").
        Tags("用户管理").
        QueryParam("page", "integer", false, "页码").
        QueryParam("size", "integer", false, "每页数量").
        SuccessResponse("成功", userListSchema).
        ErrorResponse("400", "请求参数错误").
        Build())
    
    e.Run(":8080")
}
```

访问 `http://localhost:8080/swagger` 查看API文档。

### 2. 使用注解自动生成文档

```go
// User 用户模型
type User struct {
    ID       int    `json:"id" description:"用户ID" example:"1"`
    Name     string `json:"name" description:"用户名" example:"张三"`
    Email    string `json:"email" description:"邮箱" example:"zhangsan@example.com"`
    Age      int    `json:"age" description:"年龄" example:"25"`
    IsActive bool   `json:"is_active" description:"是否激活" example:"true"`
}

// getUserByID 根据ID获取用户
// @Summary 根据ID获取用户
// @Description 通过用户ID获取单个用户的详细信息
// @Tags 用户管理
// @Param id path int true "用户ID" example(1)
// @Success 200 {object} User "成功返回用户信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Security BearerAuth
// @Router /users/{id} [get]
func getUserByID(c *gyarn.Context) {
    // 处理逻辑
}

// createUser 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body User true "用户信息"
// @Success 201 {object} User "成功创建用户"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Security BearerAuth
// @Router /users [post]
func createUser(c *gyarn.Context) {
    // 处理逻辑
}

func main() {
    e := engine.New()
    docs := openapi.EnableOpenAPI(e)
    
    // 注册模型
    openapi.RegisterModel("User", User{})
    openapi.RegisterModel("ErrorResponse", ErrorResponse{})
    
    // 从注解生成文档
    err := docs.GenerateFromAnnotations("./handlers")
    if err != nil {
        log.Fatal("生成API文档失败:", err)
    }
    
    e.GET("/users/:id", getUserByID)
    e.POST("/users", createUser)
    
    e.Run(":8080")
}
```

### 3. 使用文档构建器

```go
// 用户Schema
userSchema := openapi.NewSchemaBuilder().
    Type("object").
    Description("用户信息").
    StringProperty("name", "用户名", "张三").
    IntegerProperty("age", "年龄", 25).
    StringProperty("email", "邮箱", "user@example.com").
    BooleanProperty("is_active", "是否激活", true).
    Required("name", "email").
    Build()

// 用户列表Schema
userListSchema := openapi.ObjectSchema("用户列表响应", map[string]*openapi.Schema{
    "code":    openapi.IntegerSchema("状态码", 200),
    "message": openapi.StringSchema("消息", "success"),
    "data": openapi.ObjectSchema("数据", map[string]*openapi.Schema{
        "users": openapi.ArraySchema("用户列表", userSchema),
        "total": openapi.IntegerSchema("总数", 100),
    }),
})

// 添加到OpenAPI
docs.AddSchema("User", userSchema)
docs.AddSchema("UserListResponse", userListSchema)

// 定义API文档
apiDoc := openapi.NewDocBuilder().
    Summary("获取用户列表").
    Description("分页获取用户列表，支持搜索和排序").
    Tags("用户管理").
    QueryParam("page", "integer", false, "页码，从1开始").
    QueryParam("size", "integer", false, "每页数量，默认10").
    QueryParam("search", "string", false, "搜索关键词").
    QueryParam("sort", "string", false, "排序字段").
    HeaderParam("Authorization", "string", true, "认证令牌").
    SuccessResponse("成功", openapi.RefSchema("#/components/schemas/UserListResponse")).
    ErrorResponse("400", "请求参数错误").
    ErrorResponse("401", "未授权").
    Security("BearerAuth").
    Build()

docs.AddRoute("GET", "/users", apiDoc)
```

### 4. 安全方案配置

```go
docs := openapi.EnableOpenAPI(e)

// 添加JWT认证
docs.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer Token认证",
})

// 添加API Key认证
docs.AddSecurityScheme("ApiKeyAuth", openapi.SecurityScheme{
    Type:        "apiKey",
    In:          "header",
    Name:        "X-API-Key",
    Description: "API Key认证",
})

// 添加Basic认证
docs.AddSecurityScheme("BasicAuth", openapi.SecurityScheme{
    Type:        "http",
    Scheme:      "basic",
    Description: "HTTP Basic认证",
})

// 设置全局安全要求
docs.GetOpenAPI().Security = []openapi.SecurityRequirement{
    {"BearerAuth": []string{}},
}
```

## 支持的注解

### 基本注解

- `@Summary` - API摘要
- `@Description` - API详细描述
- `@Tags` - API标签（逗号分隔）
- `@Router` - 路由信息，格式：`/path [method]`
- `@Deprecated` - 标记为已弃用

### 参数注解

- `@Param` - 参数定义
  - 格式：`@Param name in type required "description" default(value) example(value)`
  - `in`: query, path, header, cookie
  - `type`: string, integer, number, boolean, array
  - `required`: true, false

### 请求体注解

- `@Accept` - 接受的内容类型
- 请求体通过`@Param body`定义

### 响应注解

- `@Success` - 成功响应
  - 格式：`@Success code {type} model "description"`
- `@Failure` - 失败响应
  - 格式：`@Failure code {type} model "description"`
- `@Produce` - 响应内容类型

### 安全注解

- `@Security` - 安全要求
  - 格式：`@Security scheme_name`

## 结构体标签

在Go结构体中使用以下标签来增强Schema生成：

```go
type User struct {
    ID       int    `json:"id" description:"用户ID" example:"1"`
    Name     string `json:"name" description:"用户名" example:"张三"`
    Email    string `json:"email" description:"邮箱地址" example:"user@example.com"`
    Age      int    `json:"age" description:"年龄" example:"25"`
    IsActive bool   `json:"is_active" description:"是否激活" example:"true"`
    Tags     []string `json:"tags" description:"用户标签" example:"['developer', 'admin']"`
    Profile  Profile  `json:"profile" description:"用户档案"`
    CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

type Profile struct {
    Avatar   string `json:"avatar" description:"头像URL" example:"https://example.com/avatar.jpg"`
    Bio      string `json:"bio" description:"个人简介" example:"软件开发工程师"`
    Location string `json:"location" description:"所在地" example:"北京"`
}
```

支持的标签：
- `json` - JSON字段名
- `description` - 字段描述
- `example` - 示例值

## 配置选项

```go
config := openapi.OpenAPIConfig{
    Title:       "API标题",
    Description: "API描述",
    Version:     "1.0.0",
    DocsPath:    "/swagger",    // Swagger UI路径
    JSONPath:    "/swagger/openapi.json", // OpenAPI JSON路径
    Enabled:     true,          // 是否启用
}

docs := openapi.EnableOpenAPI(engine, config)
```

## 高级功能

### 1. 自定义Swagger UI

```go
// 可以通过修改HTML模板来自定义Swagger UI
// 或者提供自己的静态文件服务
```

### 2. 多环境配置

```go
docs.AddServer(openapi.Server{
    URL:         "https://api.example.com",
    Description: "生产环境",
})

docs.AddServer(openapi.Server{
    URL:         "https://staging-api.example.com",
    Description: "测试环境",
})

docs.AddServer(openapi.Server{
    URL:         "http://localhost:8080",
    Description: "开发环境",
})
```

### 3. 外部文档链接

```go
docs.AddTag(openapi.Tag{
    Name:        "用户管理",
    Description: "用户相关的API接口",
    ExternalDocs: &openapi.ExternalDocs{
        Description: "用户管理详细文档",
        URL:         "https://docs.example.com/users",
    },
})
```

### 4. 响应示例

```go
response := openapi.Response{
    Description: "成功响应",
    Content: map[string]openapi.MediaType{
        "application/json": {
            Schema: userSchema,
            Examples: map[string]openapi.Example{
                "user1": {
                    Summary: "普通用户示例",
                    Value: map[string]interface{}{
                        "id":        1,
                        "name":      "张三",
                        "email":     "zhangsan@example.com",
                        "is_active": true,
                    },
                },
                "admin": {
                    Summary: "管理员用户示例",
                    Value: map[string]interface{}{
                        "id":        2,
                        "name":      "管理员",
                        "email":     "admin@example.com",
                        "is_active": true,
                        "role":      "admin",
                    },
                },
            },
        },
    },
}
```

## 最佳实践

1. **统一的错误响应格式**
   ```go
   type ErrorResponse struct {
       Code    int    `json:"code" description:"错误码"`
       Message string `json:"message" description:"错误信息"`
   }
   ```

2. **使用标签组织API**
   ```go
   docs.AddTag(openapi.Tag{Name: "用户管理", Description: "用户相关接口"})
   docs.AddTag(openapi.Tag{Name: "订单管理", Description: "订单相关接口"})
   ```

3. **合理使用Schema引用**
   ```go
   // 定义可复用的Schema
   docs.AddSchema("User", userSchema)
   docs.AddSchema("ErrorResponse", errorSchema)
   
   // 在API中引用
   schema := openapi.RefSchema("#/components/schemas/User")
   ```

4. **提供详细的示例**
   ```go
   .QueryParam("status", "string", false, "用户状态：active, inactive, pending")
   .Example("active")
   ```

5. **使用安全方案**
   ```go
   // 为需要认证的API添加安全要求
   .Security("BearerAuth")
   ```

## 故障排除

### 常见问题

1. **文档页面无法访问**
   - 检查DocsPath配置是否正确
   - 确认路由注册是否成功

2. **注解解析失败**
   - 检查注解格式是否正确
   - 确认源码目录路径是否正确

3. **Schema生成错误**
   - 检查结构体标签是否正确
   - 确认结构体字段是否导出

4. **Swagger UI显示异常**
   - 检查OpenAPI JSON是否有效
   - 确认网络连接是否正常（CDN资源）

### 调试技巧

1. **查看生成的OpenAPI JSON**
   ```bash
   curl http://localhost:8080/swagger/openapi.json
   ```

2. **验证OpenAPI规范**
   使用在线工具验证生成的OpenAPI文档是否符合规范。

3. **检查日志输出**
   框架会输出相关的错误和警告信息。

## 示例项目

完整的示例项目请参考 `examples/openapi-demo/` 目录。 