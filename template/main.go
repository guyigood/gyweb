package main

import (
	"fmt"
	//"github.com/guyigood/gyweb/core/openapi"
	"{project_name}/controller/sysbase"
	"{project_name}/public"
	"{project_name}/service"
	"os"
	"{project_name}/lib"
	"github.com/guyigood/gyweb/core/utils/datatype"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/openapi"
)

func main() {
	public.SysInit()
	r := engine.New()
	// 检查命令行参数，只有输入gendoc时才生成文档
	if len(os.Args) > 1 && os.Args[1] == "gendoc" {
		// 启用OpenAPI - 一行代码！
		docs := openapi.EnableOpenAPI(r, openapi.OpenAPIConfig{
			Title:       "体温计管理系统 API接口文档",
			Description: "体温计管理系统API接口说明，支持设备管理、病人管理、传感器数据统计和MQTT数据接收",
			Version:     "1.0.0",
		})

		// 从注解生成文档 - 关键的一行！
		fmt.Println("开始生成OpenAPI文档...")
		docs.GenerateFromAnnotations("./")
		docs.AutoDiscoverModels("./model", "./controller/sysbase", "./controller/dbcommon")

		fmt.Println("OpenAPI文档生成成功！")
		return
	}

	middleware.SetDebug(public.SysConfig.Server.Debug)

	// 启动MQTT服务（在后台运行）
	go func() {
		fmt.Println("正在启动MQTT服务...")
		service.StartMQTTService()
	}()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	//r.Use(lib.OptionsHandler()) // 处理OPTIONS预检请求
	r.Use(middleware.RateLimit(100))
	CustomAuth(r)      //设置为自定义鉴权
	r.Use(lib.LogDb()) // 将日志中间件放在认证中间件之后
	RegRoute(r)
	// 启动服务器
	fmt.Println("正在启动服务器，端口：" + datatype.TypetoStr(public.SysConfig.Server.Port))
	//fmt.Println("OpenAPI文档地址：http://localhost:8080/swagger")
	err := r.Run(":" + datatype.TypetoStr(public.SysConfig.Server.Port))
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

func CustomAuth(r *engine.Engine) {
	basicAuth := middleware.NewAuthManager().
		UseCustom(sysbase.CheckAuth).
		SetUnauthorizedHandler(func(c *gyarn.Context) {
			c.Error(401, "未授权")
		}).
		AddWhitelist(
			[]string{"/api/auth/login", "/api/db/build", "/api/db/clearcache", "/swagger", "/swagger/", "/swagger/index.html", "/docs", "/docs/"}, // 白名单路径
			[]string{"/static/", "/public/", "/swagger/"},                                                                                         // 白名单前缀
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
