package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/services/auth"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 创建引擎
	r := engine.New()

	// 添加基础中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 配置Redis（可选，这里使用模拟）
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// 创建鉴权服务
	authService := auth.NewAuthService(&auth.AuthServiceConfig{
		RedisClient:      redisClient,
		CachePrefix:      "example:auth:",
		CacheExpiration:  30 * time.Minute,
		EnableSuperAdmin: true,
		SuperAdminRole:   "super_admin",
		DefaultDeny:      true,
	})

	// 初始化用户和权限数据
	initializeAuthData(authService)

	// 配置JWT认证
	jwtConfig := &middleware.JWTConfig{
		SecretKey:     "your-super-secret-key-at-least-32-characters-long",
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		ExpiresIn:     24 * time.Hour,
	}

	// 公开路由（无需认证）
	r.GET("/", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "欢迎使用API鉴权示例",
			"endpoints": map[string]string{
				"登录":    "POST /api/login",
				"获取用户":  "GET /api/users (需要认证)",
				"创建用户":  "POST /api/users (需要user.create权限)",
				"删除用户":  "DELETE /api/users/:id (需要admin角色)",
				"管理员面板": "GET /api/admin/* (需要admin或super_admin角色)",
				"用户资料":  "GET /api/profile (只需认证)",
				"获取文章":  "GET /api/articles (需要article.read权限)",
				"创建文章":  "POST /api/articles (需要article.create权限)",
			},
		})
	})

	// 登录接口
	r.POST("/api/login", loginHandler(jwtConfig, authService))

	// 创建认证鉴权中间件链
	authChain := auth.CreateAuthChain(jwtConfig, authService)

	// 需要认证的API路由组
	api := r.Group("/api")
	api.Use(authChain...)

	// 只需要认证的路由
	authMiddleware := auth.NewSimpleAuthMiddleware(nil). // 只认证，不鉴权
								SkipPaths("/api/login").
								Handler()

	profile := r.Group("/api/profile")
	profile.Use(middleware.CreateJWTAuth(jwtConfig))
	profile.Use(authMiddleware)
	profile.GET("", getProfile)

	// 需要权限的路由
	api.GET("/users", getUsersHandler)

	// 创建用户（需要特定权限）
	userCreateGroup := api.Group("/users")
	userCreateGroup.Use(auth.RequirePermission("user", "create"))
	userCreateGroup.POST("", createUserHandler)

	api.GET("/users/:id", getUserByIDHandler)

	// 更新用户（需要特定权限）
	userUpdateGroup := api.Group("/users")
	userUpdateGroup.Use(auth.RequirePermission("user", "update"))
	userUpdateGroup.PUT("/:id", updateUserHandler)

	// 删除用户（需要管理员角色）
	userDeleteGroup := api.Group("/users")
	userDeleteGroup.Use(auth.RequireRole("admin", "super_admin"))
	userDeleteGroup.DELETE("/:id", deleteUserHandler)

	// 文章管理路由
	// 查看文章（需要article.read权限）
	articleReadGroup := api.Group("/articles")
	articleReadGroup.Use(auth.RequirePermission("article", "read"))
	articleReadGroup.GET("", getArticlesHandler)

	// 创建文章（需要article.create权限）
	articleCreateGroup := api.Group("/articles")
	articleCreateGroup.Use(auth.RequirePermission("article", "create"))
	articleCreateGroup.POST("", createArticleHandler)

	// 更新文章（需要article.update权限）
	articleUpdateGroup := api.Group("/articles")
	articleUpdateGroup.Use(auth.RequirePermission("article", "update"))
	articleUpdateGroup.PUT("/:id", updateArticleHandler)

	// 删除文章（需要编辑者或管理员角色）
	articleDeleteGroup := api.Group("/articles")
	articleDeleteGroup.Use(auth.RequireRole("editor", "admin"))
	articleDeleteGroup.DELETE("/:id", deleteArticleHandler)

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(auth.RequireRole("admin", "super_admin"))
	admin.GET("/dashboard", adminDashboardHandler)
	admin.GET("/logs", adminLogsHandler)
	admin.POST("/permissions", adminCreatePermissionHandler)

	// 启动服务器
	fmt.Println("🚀 服务器启动成功!")
	fmt.Println("📱 访问地址: http://localhost:8080")
	fmt.Println("📚 API文档: http://localhost:8080")
	fmt.Println("")
	fmt.Println("🔑 测试用户:")
	fmt.Println("  - 管理员: admin/password")
	fmt.Println("  - 编辑者: editor/password")
	fmt.Println("  - 普通用户: user/password")
	fmt.Println("")
	fmt.Println("🌟 测试步骤:")
	fmt.Println("1. POST /api/login 获取JWT token")
	fmt.Println("2. 在请求头中添加: Authorization: Bearer <token>")
	fmt.Println("3. 访问受保护的API端点")

	log.Fatal(r.Run(":8080"))
}

// initializeAuthData 初始化权限数据
func initializeAuthData(authService *auth.AuthService) {
	// 添加权限
	permissions := []*auth.Permission{
		{ID: "user.create", Name: "创建用户", Resource: "user", Action: "create"},
		{ID: "user.read", Name: "查看用户", Resource: "user", Action: "read"},
		{ID: "user.update", Name: "更新用户", Resource: "user", Action: "update"},
		{ID: "user.delete", Name: "删除用户", Resource: "user", Action: "delete"},
		{ID: "article.create", Name: "创建文章", Resource: "article", Action: "create"},
		{ID: "article.read", Name: "查看文章", Resource: "article", Action: "read"},
		{ID: "article.update", Name: "更新文章", Resource: "article", Action: "update"},
		{ID: "article.delete", Name: "删除文章", Resource: "article", Action: "delete"},
	}

	// 添加角色
	roles := []*auth.Role{
		{
			ID:   "user",
			Name: "普通用户",
			Permissions: []auth.Permission{
				*permissions[1], // user.read
				*permissions[5], // article.read
			},
		},
		{
			ID:   "editor",
			Name: "编辑者",
			Permissions: []auth.Permission{
				*permissions[1], // user.read
				*permissions[4], // article.create
				*permissions[5], // article.read
				*permissions[6], // article.update
			},
			ParentRoles: []string{"user"},
		},
		{
			ID:   "admin",
			Name: "管理员",
			Permissions: []auth.Permission{
				*permissions[0], // user.create
				*permissions[1], // user.read
				*permissions[2], // user.update
				*permissions[3], // user.delete
			},
			ParentRoles: []string{"editor"},
		},
	}

	for _, role := range roles {
		authService.AddRole(role)
	}

	// 添加测试用户
	users := []*auth.User{
		{
			ID:       "admin001",
			Username: "admin",
			Roles:    []string{"admin"},
			Attributes: map[string]string{
				"password":   "password",
				"department": "tech",
				"level":      "senior",
			},
		},
		{
			ID:       "editor001",
			Username: "editor",
			Roles:    []string{"editor"},
			Attributes: map[string]string{
				"password":   "password",
				"department": "content",
				"level":      "junior",
			},
		},
		{
			ID:       "user001",
			Username: "user",
			Roles:    []string{"user"},
			Attributes: map[string]string{
				"password":   "password",
				"department": "sales",
				"level":      "entry",
			},
		},
	}

	for _, user := range users {
		authService.AddUser(user)
	}

	// 添加权限规则
	rules := []*auth.AuthorizationRule{
		{
			ID:         "public_read_access",
			Resource:   "*",
			Action:     "read",
			Method:     "GET",
			Path:       "^/api/public/.*",
			AllowRoles: []string{"*"},
			Priority:   100,
			IsActive:   true,
		},
		{
			ID:         "admin_full_access",
			Resource:   "*",
			Action:     "*",
			Method:     "*",
			Path:       "/api/admin/.*",
			AllowRoles: []string{"admin", "super_admin"},
			Priority:   1000,
			IsActive:   true,
		},
		{
			ID:       "user_own_data_access",
			Resource: "user",
			Action:   "read",
			Method:   "GET",
			Path:     "/api/users/\\d+",
			Conditions: map[string]string{
				"owner_only": "true",
			},
			AllowRoles: []string{"user"},
			Priority:   500,
			IsActive:   true,
		},
	}

	for _, rule := range rules {
		authService.AddRule(rule)
	}
}

// loginHandler 登录处理器
func loginHandler(jwtConfig *middleware.JWTConfig, authService *auth.AuthService) gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		var loginReq struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 验证用户密码（这里简化处理）
		user, err := validateUser(authService, loginReq.Username, loginReq.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "用户名或密码错误",
			})
			return
		}

		// 生成JWT Token
		role := ""
		if len(user.Roles) > 0 {
			role = user.Roles[0]
		}

		token, err := middleware.GenerateJWT(jwtConfig, user.ID, user.Username, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "生成Token失败",
			})
			return
		}

		// 获取用户权限
		permissions, _ := authService.GetUserPermissions(user.ID)

		c.JSON(http.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "登录成功",
			"data": map[string]interface{}{
				"token":       token,
				"user":        user,
				"permissions": permissions,
			},
		})
	}
}

// validateUser 验证用户密码
func validateUser(authService *auth.AuthService, username, password string) (*auth.User, error) {
	// 简化实现：根据用户名查找用户
	userIDs := []string{"admin001", "editor001", "user001"}
	usernames := []string{"admin", "editor", "user"}

	for i, uname := range usernames {
		if uname == username {
			user, err := authService.GetUser(userIDs[i])
			if err != nil {
				return nil, err
			}

			// 验证密码
			if user.Attributes["password"] == password {
				return user, nil
			}
		}
	}

	return nil, fmt.Errorf("用户名或密码错误")
}

// API处理器函数
func getProfile(c *gyarn.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取用户资料成功",
		"data": map[string]interface{}{
			"user_id":  userID,
			"username": username,
		},
	})
}

func getUsersHandler(c *gyarn.Context) {
	authCtx, exists := auth.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "未找到认证信息",
		})
		return
	}

	// 模拟获取用户列表
	users := []map[string]interface{}{
		{"id": "1", "username": "admin", "role": "admin"},
		{"id": "2", "username": "editor", "role": "editor"},
		{"id": "3", "username": "user", "role": "user"},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取用户列表成功",
		"data": map[string]interface{}{
			"users":     users,
			"requester": authCtx.UserID,
		},
	})
}

func createUserHandler(c *gyarn.Context) {
	var userReq struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	if err := c.BindJSON(&userReq); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "创建用户成功",
		"data": map[string]interface{}{
			"id":       "new_user_" + strconv.FormatInt(time.Now().Unix(), 10),
			"username": userReq.Username,
			"role":     userReq.Role,
		},
	})
}

func getUserByIDHandler(c *gyarn.Context) {
	userID := c.Param("id")
	authCtx, _ := auth.GetAuthContext(c)

	// 检查是否访问自己的数据
	if userID == authCtx.UserID || auth.HasRole(c, "admin") {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取用户信息成功",
			"data": map[string]interface{}{
				"id":       userID,
				"username": "sample_user",
				"role":     "user",
			},
		})
	} else {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    403,
			"message": "只能查看自己的用户信息",
		})
	}
}

func updateUserHandler(c *gyarn.Context) {
	userID := c.Param("id")

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "更新用户成功",
		"data": map[string]interface{}{
			"id": userID,
		},
	})
}

func deleteUserHandler(c *gyarn.Context) {
	userID := c.Param("id")

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "删除用户成功",
		"data": map[string]interface{}{
			"id": userID,
		},
	})
}

func getArticlesHandler(c *gyarn.Context) {
	articles := []map[string]interface{}{
		{"id": "1", "title": "Go语言入门", "author": "admin"},
		{"id": "2", "title": "Web开发实践", "author": "editor"},
		{"id": "3", "title": "API设计指南", "author": "user"},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取文章列表成功",
		"data":    articles,
	})
}

func createArticleHandler(c *gyarn.Context) {
	var articleReq struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&articleReq); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "创建文章成功",
		"data": map[string]interface{}{
			"id":     "article_" + strconv.FormatInt(time.Now().Unix(), 10),
			"title":  articleReq.Title,
			"author": authCtx.Username,
		},
	})
}

func updateArticleHandler(c *gyarn.Context) {
	articleID := c.Param("id")

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "更新文章成功",
		"data": map[string]interface{}{
			"id": articleID,
		},
	})
}

func deleteArticleHandler(c *gyarn.Context) {
	articleID := c.Param("id")

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "删除文章成功",
		"data": map[string]interface{}{
			"id": articleID,
		},
	})
}

func adminDashboardHandler(c *gyarn.Context) {
	authCtx, _ := auth.GetAuthContext(c)

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取管理员面板数据成功",
		"data": map[string]interface{}{
			"admin":    authCtx.Username,
			"users":    125,
			"articles": 450,
			"today_pv": 1234,
		},
	})
}

func adminLogsHandler(c *gyarn.Context) {
	logs := []map[string]interface{}{
		{"time": "2025-01-20 10:30:00", "user": "user001", "action": "login", "result": "success"},
		{"time": "2025-01-20 10:35:00", "user": "editor001", "action": "create_article", "result": "success"},
		{"time": "2025-01-20 10:40:00", "user": "user002", "action": "login", "result": "failed"},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取日志成功",
		"data":    logs,
	})
}

func adminCreatePermissionHandler(c *gyarn.Context) {
	var permReq struct {
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	if err := c.BindJSON(&permReq); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "创建权限成功",
		"data": map[string]interface{}{
			"id":          permReq.Resource + "." + permReq.Action,
			"resource":    permReq.Resource,
			"action":      permReq.Action,
			"description": permReq.Description,
		},
	})
}
