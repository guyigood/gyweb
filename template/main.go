package main

import (
	"fmt"
	"net/http"
	"{project_name}/controller/sysbase"
	"{project_name}/lib"
	"{project_name}/public"

	_ "github.com/go-sql-driver/mysql"
	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/openapi"
	orm "github.com/guyigood/gyweb/core/orm/mysql"
)

func main() {
	public.SysInit()
	r := engine.New()
	// 启用OpenAPI - 一行代码！
	docs := openapi.EnableOpenAPI(r, openapi.OpenAPIConfig{
		Title:       "FirstWeb API接口文档",
		Description: "FirstWeb项目API接口说明，支持用户认证和数据库通用操作",
		Version:     "1.0.0",
	})

	// 从注解生成文档 - 关键的一行！
	fmt.Println("开始生成OpenAPI文档...")
	err := docs.GenerateFromAnnotations("./")

	middleware.SetDebug(public.SysConfig.Server.Debug)

	if err != nil {
		fmt.Printf("OpenApi 引擎生成失败！错误: %v\n", err)
		return
	}
	docs.GenerateFromAnnotations("./controller/sysbase")
	docs.GenerateFromAnnotations("./controller/dbcommon")

	fmt.Println("OpenAPI文档生成成功！")
	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100))
	CustomAuth(r) //设置为自定义鉴权
	RegRoute(r)
	r.Use(lib.LogDb())
	// 启动服务器
	fmt.Println("正在启动服务器，端口：8080")
	fmt.Println("OpenAPI文档地址：http://localhost:8080/swagger")
	err = r.Run(":8080")
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

func CustomAuth(r *engine.Engine) {
	basicAuth := middleware.NewAuthManager().
		UseCustom(sysbase.CheckAuth).
		SetUnauthorizedHandler(func(c *gyarn.Context) {
			c.Error(103, "未授权")
		}).
		AddWhitelist(
			[]string{"/api/auth/login", "/swagger", "/swagger/", "/swagger/index.html", "/docs", "/docs/"}, // 白名单路径
			[]string{"/static/", "/public/", "/swagger/"},                                                  // 白名单前缀
			nil,
		).
		Build()
	r.Use(basicAuth)

}
func JwtAuth(r *engine.Engine) {
	jwtAuth := middleware.NewAuthManager().
		UseJWT(public.GetJwtConfig()).
		AddWhitelist(
			[]string{"/api/login", "/api/register", "/api/dbtest"}, // 白名单路径
			[]string{"/static/", "/public/"},                       // 白名单前缀
			nil,
		).
		Build()
	r.Use(jwtAuth)
	{
		r.POST("/api/login", sysbase.Login)
	}
}
func Login(c *gyarn.Context) {
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
		role := "admin"
		token, err := middleware.GenerateJWT(public.GetJwtConfig(), "1", loginForm.Username, role)
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
}

func RegTest(r *engine.Engine) {
	// 注册路由
	r.GET("/", func(c *gyarn.Context) {
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Welcome to GyWeb!",
		})
	})

	// 路由组
	api := r.Group("/api")
	{
		api.GET("/users", func(c *gyarn.Context) {
			c.JSON(http.StatusOK, gyarn.H{
				"users": []string{"Alice", "Bob"},
			})
		})
		api.GET("/dbtest", func(c *gyarn.Context) {
			db, err := orm.NewDB("mysql", "root:gy7210@tcp(localhost:3306)/cpnrc?charset=utf8mb4&parseTime=True&loc=Local")
			if err != nil {
				fmt.Println(err)
				return
			}

			data, _ := db.Table("sl_agv_call").Where("inc_type=?", "in").Limit(10).Offset(20).All()

			c.Success(data)
		})
	}
}
