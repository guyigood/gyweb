package main

import (
	"log"
	"strconv"
	"time"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/openapi"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" description:"用户ID" example:"1"`
	Name      string    `json:"name" description:"用户�? example:"张三"`
	Email     string    `json:"email" description:"邮箱地址" example:"zhangsan@example.com"`
	Age       int       `json:"age" description:"年龄" example:"25"`
	IsActive  bool      `json:"is_active" description:"是否激�? example:"true"`
	Tags      []string  `json:"tags" description:"用户标签"`
	Profile   Profile   `json:"profile" description:"用户档案"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

// Profile 用户档案
type Profile struct {
	Avatar   string `json:"avatar" description:"头像URL" example:"https://example.com/avatar.jpg"`
	Bio      string `json:"bio" description:"个人简�? example:"软件开发工程师"`
	Location string `json:"location" description:"所在地" example:"北京"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name    string   `json:"name" description:"用户�? example:"张三"`
	Email   string   `json:"email" description:"邮箱地址" example:"zhangsan@example.com"`
	Age     int      `json:"age" description:"年龄" example:"25"`
	Tags    []string `json:"tags" description:"用户标签"`
	Profile Profile  `json:"profile" description:"用户档案"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name     *string  `json:"name,omitempty" description:"用户�? example:"李四"`
	Email    *string  `json:"email,omitempty" description:"邮箱地址" example:"lisi@example.com"`
	Age      *int     `json:"age,omitempty" description:"年龄" example:"30"`
	IsActive *bool    `json:"is_active,omitempty" description:"是否激�? example:"false"`
	Tags     []string `json:"tags,omitempty" description:"用户标签"`
	Profile  *Profile `json:"profile,omitempty" description:"用户档案"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Code    int    `json:"code" description:"状态码" example:"200"`
	Message string `json:"message" description:"响应消息" example:"success"`
	Data    struct {
		Users []User `json:"users" description:"用户列表"`
		Total int    `json:"total" description:"总数" example:"100"`
		Page  int    `json:"page" description:"当前页码" example:"1"`
		Size  int    `json:"size" description:"每页数量" example:"10"`
	} `json:"data" description:"响应数据"`
}

// UserResponse 单个用户响应
type UserResponse struct {
	Code    int    `json:"code" description:"状态码" example:"200"`
	Message string `json:"message" description:"响应消息" example:"success"`
	Data    User   `json:"data" description:"用户信息"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code" description:"错误�? example:"400"`
	Message string `json:"message" description:"错误信息" example:"请求参数错误"`
}

// 模拟数据存储
var users = []User{
	{
		ID:       1,
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      25,
		IsActive: true,
		Tags:     []string{"developer", "golang"},
		Profile: Profile{
			Avatar:   "https://example.com/avatar1.jpg",
			Bio:      "Go语言开发工程师",
			Location: "北京",
		},
		CreatedAt: time.Now().AddDate(0, -1, 0),
	},
	{
		ID:       2,
		Name:     "李四",
		Email:    "lisi@example.com",
		Age:      30,
		IsActive: true,
		Tags:     []string{"manager", "product"},
		Profile: Profile{
			Avatar:   "https://example.com/avatar2.jpg",
			Bio:      "产品经理",
			Location: "上海",
		},
		CreatedAt: time.Now().AddDate(0, -2, 0),
	},
}

var nextID = 3

// getUserList 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持搜索和排序
// @Tags 用户管理
// @Param page query int false "页码，从1开�? default(1) example(1)
// @Param size query int false "每页数量，默�?0" default(10) example(10)
// @Param search query string false "搜索关键�? example("张三")
// @Param sort query string false "排序字段：id, name, email, age, created_at" example("id")
// @Param order query string false "排序方向：asc, desc" default(asc) example("asc")
// @Success 200 {object} UserListResponse "成功返回用户列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Security BearerAuth
// @Router /users [get]
func getUserList(c *gyarn.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	search := c.Query("search")
	//sort := c.DefaultQuery("sort", "id")
	//order := c.DefaultQuery("order", "asc")

	// 参数验证
	if page < 1 {
		c.JSON(400, ErrorResponse{Code: 400, Message: "页码必须大于0"})
		return
	}
	if size < 1 || size > 100 {
		c.JSON(400, ErrorResponse{Code: 400, Message: "每页数量必须�?-100之间"})
		return
	}

	// 模拟搜索和排序逻辑
	filteredUsers := users
	if search != "" {
		var filtered []User
		for _, user := range users {
			if user.Name == search || user.Email == search {
				filtered = append(filtered, user)
			}
		}
		filteredUsers = filtered
	}

	// 模拟分页
	total := len(filteredUsers)
	start := (page - 1) * size
	end := start + size
	if start >= total {
		filteredUsers = []User{}
	} else if end > total {
		filteredUsers = filteredUsers[start:]
	} else {
		filteredUsers = filteredUsers[start:end]
	}

	response := UserListResponse{
		Code:    200,
		Message: "success",
	}
	response.Data.Users = filteredUsers
	response.Data.Total = total
	response.Data.Page = page
	response.Data.Size = size

	c.JSON(200, response)
}

// getUserByID 根据ID获取用户
// @Summary 根据ID获取用户
// @Description 通过用户ID获取单个用户的详细信�?
// @Tags 用户管理
// @Param id path int true "用户ID" example(1)
// @Success 200 {object} UserResponse "成功返回用户信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存�?
// @Security BearerAuth
// @Router /users/{id} [get]
func getUserByID(c *gyarn.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, ErrorResponse{Code: 400, Message: "用户ID必须是数"})
		return
	}

	// 查找用户
	for _, user := range users {
		if user.ID == id {
			c.JSON(200, UserResponse{
				Code:    200,
				Message: "success",
				Data:    user,
			})
			return
		}
	}

	c.JSON(404, ErrorResponse{Code: 404, Message: "用户不存"})
}

// createUser 创建用户
// @Summary 创建新用
// @Description 创建一个新的用户账
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} UserResponse "成功创建用户"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Security BearerAuth
// @Router /users [post]
func createUser(c *gyarn.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrorResponse{Code: 400, Message: "请求参数格式错误"})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Email == "" {
		c.JSON(400, ErrorResponse{Code: 400, Message: "用户名和邮箱不能为空"})
		return
	}

	// 检查邮箱是否已存在
	for _, user := range users {
		if user.Email == req.Email {
			c.JSON(400, ErrorResponse{Code: 400, Message: "邮箱已存在"})
			return
		}
	}

	//创建新用户
	newUser := User{
		ID:        nextID,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		IsActive:  true,
		Tags:      req.Tags,
		Profile:   req.Profile,
		CreatedAt: time.Now(),
	}
	nextID++

	users = append(users, newUser)

	c.JSON(201, UserResponse{
		Code:    201,
		Message: "用户创建成功",
		Data:    newUser,
	})
}

// updateUser 更新用户
// @Summary 更新用户信息
// @Description 更新指定用户的信息，支持部分更新
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID" example(1)
// @Param user body UpdateUserRequest true "更新的用户信"
// @Success 200 {object} UserResponse "成功更新用户"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存"
// @Security BearerAuth
// @Router /users/{id} [put]
func updateUser(c *gyarn.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, ErrorResponse{Code: 400, Message: "用户ID必须是数"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrorResponse{Code: 400, Message: "请求参数格式错误"})
		return
	}

	// 查找并更新用�?
	for i, user := range users {
		if user.ID == id {
			if req.Name != nil {
				users[i].Name = *req.Name
			}
			if req.Email != nil {
				// 检查邮箱是否已被其他用户使�?
				for _, otherUser := range users {
					if otherUser.ID != id && otherUser.Email == *req.Email {
						c.JSON(400, ErrorResponse{Code: 400, Message: "邮箱已被其他用户使用"})
						return
					}
				}
				users[i].Email = *req.Email
			}
			if req.Age != nil {
				users[i].Age = *req.Age
			}
			if req.IsActive != nil {
				users[i].IsActive = *req.IsActive
			}
			if req.Tags != nil {
				users[i].Tags = req.Tags
			}
			if req.Profile != nil {
				users[i].Profile = *req.Profile
			}

			c.JSON(200, UserResponse{
				Code:    200,
				Message: "用户更新成功",
				Data:    users[i],
			})
			return
		}
	}

	c.JSON(404, ErrorResponse{Code: 404, Message: "用户不存"})
}

// deleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Param id path int true "用户ID" example(1)
// @Success 200 {object} ErrorResponse "成功删除用户"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存"
// @Security BearerAuth
// @Router /users/{id} [delete]
func deleteUser(c *gyarn.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, ErrorResponse{Code: 400, Message: "用户ID必须是数"})
		return
	}

	// 查找并删除用�?
	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			c.JSON(200, ErrorResponse{Code: 200, Message: "用户删除成功"})
			return
		}
	}

	c.JSON(404, ErrorResponse{Code: 404, Message: "用户不存"})
}

// getHealth 健康检�?
// @Summary 健康检�?
// @Description 检查服务是否正常运�?
// @Tags 系统
// @Success 200 {object} ErrorResponse "服务正常"
// @Router /health [get]
func getHealth(c *gyarn.Context) {
	c.JSON(200, ErrorResponse{Code: 200, Message: "服务正常运行"})
}

func main() {
	// 创建引擎
	e := engine.New()

	// 启用OpenAPI支持
	docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
		Title:       "GyWeb API 示例",
		Description: "这是一个使用GyWeb框架构建的API示例，展示了OpenAPI文档生成功能",
		Version:     "1.0.0",
		DocsPath:    "/swagger",
	})

	// 添加服务器信�?
	docs.AddServer(openapi.Server{
		URL:         "http://localhost:8082",
		Description: "开发环",
	})

	docs.AddServer(openapi.Server{
		URL:         "https://api.example.com",
		Description: "生产环境",
	})

	// 添加标签
	docs.AddTag(openapi.Tag{
		Name:        "用户管理",
		Description: "用户相关的API接口，包括增删改查操",
		ExternalDocs: &openapi.ExternalDocs{
			Description: "用户管理详细文档",
			URL:         "https://docs.example.com/users",
		},
	})

	docs.AddTag(openapi.Tag{
		Name:        "系统",
		Description: "系统相关的API接口",
	})

	// 注册模型
	openapi.RegisterModel("User", User{})
	openapi.RegisterModel("Profile", Profile{})
	openapi.RegisterModel("CreateUserRequest", CreateUserRequest{})
	openapi.RegisterModel("UpdateUserRequest", UpdateUserRequest{})
	openapi.RegisterModel("UserListResponse", UserListResponse{})
	openapi.RegisterModel("UserResponse", UserResponse{})
	openapi.RegisterModel("ErrorResponse", ErrorResponse{})

	// 从注解生成文�?
	err := docs.GenerateFromAnnotations("./")
	if err != nil {
		log.Printf("生成API文档失败: %v", err)
	}

	// 手动添加一些额外的API文档（演示手动方式）
	docs.AddRoute("GET", "/users", openapi.NewDocBuilder().
		Summary("获取用户列表").
		Description("分页获取用户列表，支持搜索和排序功能").
		Tags("用户管理").
		QueryParam("page", "integer", false, "页码，从1开").
		QueryParam("size", "integer", false, "每页数量，默0，最00").
		QueryParam("search", "string", false, "搜索关键词，支持按用户名或邮箱搜").
		QueryParam("sort", "string", false, "排序字段：id, name, email, age, created_at").
		QueryParam("order", "string", false, "排序方向：asc（升序）, desc（降序）").
		SuccessResponse("成功", openapi.RefSchema("#/components/schemas/UserListResponse")).
		ErrorResponse("400", "请求参数错误").
		Security("BearerAuth").
		Build())

	// 注册路由
	e.GET("/health", getHealth)
	e.GET("/users", getUserList)
	e.GET("/users/:id", getUserByID)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// 启动服务�?
	log.Println("服务器启动在 http://localhost:8082")
	log.Println("API文档地址: http://localhost:8082/swagger")
	log.Println("OpenAPI JSON: http://localhost:8082/swagger/openapi.json")

	if err := e.Run(":8082"); err != nil {
		log.Fatal("服务器启动失", err)
	}
}
