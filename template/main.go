package main

import (
	"fmt"
	"{project_name}/controller/sysbase"
	"{project_name}/lib"
	"{project_name}/public"

	_ "github.com/go-sql-driver/mysql"
	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/openapi"
)

func main() {
	public.SysInit()
	r := engine.New()
	// 启用OpenAPI - 一行代码！
	docs := openapi.EnableOpenAPI(r, openapi.OpenAPIConfig{
		Title:       "体温计管理系统 API接口文档",
		Description: "体温计管理系统API接口说明，支持设备管理、病人管理、传感器数据统计和MQTT数据接收",
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
	docs.GenerateFromAnnotations("./controller/statistics")

	fmt.Println("OpenAPI文档生成成功！")

	// 启动MQTT服务（在后台运行）
	go func() {
		fmt.Println("正在启动MQTT服务...")
		service.StartMQTTService()
	}()

	// 使用中间件 - 注意顺序很重要
	r.Use(middleware.CORS()) // CORS必须在最前面
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimit(100))
	CustomAuth(r)      //设置为自定义鉴权
	r.Use(lib.LogDb()) // 将日志中间件放在认证中间件之后

	// 添加测试路由
	AddTestRoutes(r)

	RegRoute(r)
	// 启动服务器
	fmt.Println("正在启动服务器，端口：8080")
	fmt.Println("OpenAPI文档地址：http://localhost:8080/swagger")
	fmt.Println("CORS测试页面：请在浏览器中打开 cors_test.html")
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
			[]string{
				"/api/auth/login",
				"/api/test",      // 添加测试路径
				"/debug/request", // 添加调试路径
				"/cors/status",   // 添加CORS状态检查
				"/swagger", "/swagger/", "/swagger/index.html",
				"/docs", "/docs/",
			}, // 白名单路径
			[]string{"/static/", "/public/", "/swagger/", "/api/test/"}, // 白名单前缀
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
