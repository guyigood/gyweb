# GyWeb OpenAPI 示例项目

这是一个完整的GyWeb OpenAPI功能演示项目，展示了如何使用GyWeb框架的OpenAPI文档生成功能。

## 功能特性

- ✅ **完整的CRUD API** - 用户管理的增删改查操作
- ✅ **OpenAPI 3.0.3文档** - 自动生成的API文档
- ✅ **Swagger UI集成** - 美观的交互式文档界面
- ✅ **注解驱动** - 通过注释自动生成文档
- ✅ **Schema自动生成** - 从Go结构体自动生成OpenAPI Schema
- ✅ **多种认证方案** - JWT Bearer Token认证示例
- ✅ **参数验证** - 完整的请求参数验证
- ✅ **错误处理** - 统一的错误响应格式

## 快速开始

### 1. 运行项目

```bash
# 进入项目目录
cd examples/openapi-demo

# 运行项目
go run main.go
```

### 2. 访问文档

启动后访问以下地址：

- **Swagger UI文档**: http://localhost:8080/swagger
- **OpenAPI JSON**: http://localhost:8080/swagger/openapi.json
- **健康检查**: http://localhost:8080/health

## API接口

### 用户管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/users` | 获取用户列表（支持分页、搜索、排序） |
| GET | `/users/{id}` | 根据ID获取用户详情 |
| POST | `/users` | 创建新用户 |
| PUT | `/users/{id}` | 更新用户信息 |
| DELETE | `/users/{id}` | 删除用户 |

### 系统接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |

## 数据模型

### User（用户）
```json
{
  "id": 1,
  "name": "张三",
  "email": "zhangsan@example.com",
  "age": 25,
  "is_active": true,
  "tags": ["developer", "golang"],
  "profile": {
    "avatar": "https://example.com/avatar.jpg",
    "bio": "Go语言开发工程师",
    "location": "北京"
  },
  "created_at": "2023-01-01T00:00:00Z"
}
```

### CreateUserRequest（创建用户请求）
```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "age": 25,
  "tags": ["developer", "golang"],
  "profile": {
    "avatar": "https://example.com/avatar.jpg",
    "bio": "Go语言开发工程师",
    "location": "北京"
  }
}
```

### UpdateUserRequest（更新用户请求）
```json
{
  "name": "李四",
  "email": "lisi@example.com",
  "age": 30,
  "is_active": false,
  "tags": ["manager", "product"],
  "profile": {
    "avatar": "https://example.com/avatar2.jpg",
    "bio": "产品经理",
    "location": "上海"
  }
}
```

## 使用示例

### 1. 获取用户列表

```bash
curl -X GET "http://localhost:8080/users?page=1&size=10&search=张三&sort=id&order=asc"
```

### 2. 获取单个用户

```bash
curl -X GET "http://localhost:8080/users/1"
```

### 3. 创建用户

```bash
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "王五",
    "email": "wangwu@example.com",
    "age": 28,
    "tags": ["designer", "ui"],
    "profile": {
      "avatar": "https://example.com/avatar3.jpg",
      "bio": "UI设计师",
      "location": "深圳"
    }
  }'
```

### 4. 更新用户

```bash
curl -X PUT "http://localhost:8080/users/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三（已更新）",
    "age": 26
  }'
```

### 5. 删除用户

```bash
curl -X DELETE "http://localhost:8080/users/1"
```

## 项目结构

```
examples/openapi-demo/
├── main.go          # 主程序文件
├── README.md        # 项目说明
└── go.mod          # Go模块文件（如果需要）
```

## 代码特点

### 1. 注解驱动的文档生成

```go
// getUserByID 根据ID获取用户
// @Summary 根据ID获取用户
// @Description 通过用户ID获取单个用户的详细信息
// @Tags 用户管理
// @Param id path int true "用户ID" example(1)
// @Success 200 {object} UserResponse "成功返回用户信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Security BearerAuth
// @Router /users/{id} [get]
func getUserByID(c *gyarn.Context) {
    // 实现逻辑
}
```

### 2. 结构体标签增强

```go
type User struct {
    ID       int    `json:"id" description:"用户ID" example:"1"`
    Name     string `json:"name" description:"用户名" example:"张三"`
    Email    string `json:"email" description:"邮箱地址" example:"zhangsan@example.com"`
    Age      int    `json:"age" description:"年龄" example:"25"`
    IsActive bool   `json:"is_active" description:"是否激活" example:"true"`
}
```

### 3. 流式API构建

```go
docs.AddRoute("GET", "/users", openapi.NewDocBuilder().
    Summary("获取用户列表").
    Description("分页获取用户列表，支持搜索和排序功能").
    Tags("用户管理").
    QueryParam("page", "integer", false, "页码，从1开始").
    QueryParam("size", "integer", false, "每页数量，默认10，最大100").
    SuccessResponse("成功", openapi.RefSchema("#/components/schemas/UserListResponse")).
    ErrorResponse("400", "请求参数错误").
    Security("BearerAuth").
    Build())
```

## 扩展功能

### 1. 添加认证中间件

```go
// JWT认证中间件示例
func authMiddleware() middleware.HandlerFunc {
    return func(c *gyarn.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, ErrorResponse{Code: 401, Message: "缺少认证令牌"})
            c.Abort()
            return
        }
        // 验证JWT令牌逻辑
        c.Next()
    }
}

// 应用到需要认证的路由
userGroup := e.Group("/users")
userGroup.Use(authMiddleware())
```

### 2. 添加更多API端点

```go
// 批量操作
e.POST("/users/batch", batchCreateUsers)
e.PUT("/users/batch", batchUpdateUsers)
e.DELETE("/users/batch", batchDeleteUsers)

// 统计接口
e.GET("/users/stats", getUserStats)
```

### 3. 添加文件上传

```go
// @Summary 上传用户头像
// @Description 上传用户头像图片
// @Tags 用户管理
// @Accept multipart/form-data
// @Param id path int true "用户ID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} UserResponse "上传成功"
// @Router /users/{id}/avatar [post]
func uploadAvatar(c *gyarn.Context) {
    // 文件上传逻辑
}
```

## 最佳实践

1. **统一响应格式** - 所有API都使用统一的响应结构
2. **详细的注解** - 为每个API提供完整的注解信息
3. **参数验证** - 对所有输入参数进行验证
4. **错误处理** - 提供清晰的错误信息和状态码
5. **示例数据** - 在Schema中提供示例值
6. **标签分组** - 使用标签对API进行逻辑分组
7. **安全方案** - 为需要认证的API添加安全要求

## 故障排除

### 常见问题

1. **文档无法访问**
   - 检查服务是否正常启动
   - 确认端口8080没有被占用

2. **注解解析失败**
   - 检查注解格式是否正确
   - 确认函数注释格式符合规范

3. **Schema显示异常**
   - 检查结构体标签是否正确
   - 确认字段是否为导出字段（首字母大写）

### 调试技巧

1. **查看生成的OpenAPI JSON**
   ```bash
   curl http://localhost:8080/swagger/openapi.json | jq
   ```

2. **检查服务日志**
   ```bash
   go run main.go
   ```

3. **验证API响应**
   ```bash
   curl -v http://localhost:8080/health
   ```

## 更多资源

- [GyWeb框架文档](../../README.md)
- [OpenAPI规范](https://swagger.io/specification/)
- [Swagger UI文档](https://swagger.io/tools/swagger-ui/)
- [Go结构体标签](https://golang.org/ref/spec#Struct_types) 