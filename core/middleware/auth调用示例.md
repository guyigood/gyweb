```go
package main

import (
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

func main() {
	r := engine.New()

	// 1. JWT认证示例
	jwtConfig := &middleware.JWTConfig{
		SecretKey:     "your-secret-key",           // JWT密钥
		TokenLookup:   "header:Authorization",      // 从请求头获取token
		TokenHeadName: "Bearer",                    // token前缀
		ExpiresIn:     24 * time.Hour,              // token过期时间
	}

	// 创建JWT认证中间件
	jwtAuth := middleware.NewAuthManager().
		UseJWT(jwtConfig).
		AddWhitelist(
			[]string{"/api/login", "/api/register"}, // 白名单路径
			[]string{"/static/", "/public/"},        // 白名单前缀
			nil,
		).
		Build()

	// 2. Session认证示例
	sessionConfig := &middleware.SessionConfig{
		SecretKey: "your-session-secret", // Session密钥
		MaxAge:    86400,                 // Session过期时间（秒）
		Path:      "/",                   // Cookie路径
		Domain:    "",                    // Cookie域名
		Secure:    false,                 // 是否只在HTTPS下传输
		HttpOnly:  true,                  // 是否禁止JavaScript访问
	}

	// 创建Session认证中间件
	sessionAuth := middleware.NewAuthManager().
		UseSession(sessionConfig).
		AddWhitelist(
			[]string{"/login", "/register"},
			[]string{"/static/"},
			nil,
		).
		Build()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// API路由组（使用JWT认证）
	api := r.Group("/api")
	api.Use(jwtAuth)
	{
		// 登录接口
		api.POST("/login", func(c *gyarn.Context) {
			var loginForm struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.BindJSON(&loginForm); err != nil {
				c.BadRequest("无效的请求参数")
				return
			}

			// 验证用户名密码（示例）
			if loginForm.Username == "admin" && loginForm.Password == "123456" {
				// 生成JWT token
                role:="admin"
				token, err := middleware.GenerateJWT(jwtConfig, 1, loginForm.Username,role)
				if err != nil {
					c.InternalServerError("生成token失败")
					return
				}

				c.Success(gyarn.H{
					"token": token,
					"user": gyarn.H{
						"id":       1,
						"username": loginForm.Username,
					},
				})
				return
			}

			c.Unauthorized("用户名或密码错误")
		})

		// 获取用户信息接口
		api.GET("/user", func(c *gyarn.Context) {
			// 从上下文中获取用户信息（由中间件设置）
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")

			c.Success(gyarn.H{
				"id":       userID,
				"username": username,
			})
		})
	}

	// Web路由组（使用Session认证）
	web := r.Group("")
	web.Use(sessionAuth)
	{
		// 登录页面
		web.GET("/login", func(c *gyarn.Context) {
			c.HTML(http.StatusOK, "<h1>登录页面</h1>")
		})

		// 登录处理
		web.POST("/login", func(c *gyarn.Context) {
			var loginForm struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.BindJSON(&loginForm); err != nil {
				c.BadRequest("无效的请求参数")
				return
			}

			// 验证用户名密码（示例）
			if loginForm.Username == "admin" && loginForm.Password == "123456" {
				// 设置Session
				err := middleware.SetSession(c, sessionConfig, 1, loginForm.Username)
				if err != nil {
					c.InternalServerError("设置session失败")
					return
				}

				c.Success(gyarn.H{
					"message": "登录成功",
					"user": gyarn.H{
						"id":       1,
						"username": loginForm.Username,
					},
				})
				return
			}

			c.Unauthorized("用户名或密码错误")
		})

		// 用户主页
		web.GET("/home", func(c *gyarn.Context) {
			// 从上下文中获取用户信息（由中间件设置）
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")

			c.HTML(http.StatusOK, fmt.Sprintf(`
				<h1>欢迎, %s!</h1>
				<p>用户ID: %d</p>
				<a href="/logout">退出登录</a>
			`, username, userID))
		})

		// 退出登录
		web.GET("/logout", func(c *gyarn.Context) {
			if err := middleware.ClearSession(c); err != nil {
				c.InternalServerError("清除session失败")
				return
			}
			c.Success(gyarn.H{"message": "已退出登录"})
		})
	}

	// 启动服务器
	r.Run(":8080")
}
```