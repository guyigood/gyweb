package main

import (
	"log"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/openapi"
)

// 数据模型
type User struct {
	ID   int    `json:"id" description:"用户ID" example:"1"`
	Name string `json:"name" description:"用户名" example:"张三"`
}

// getHealth 健康检查
// @Summary 健康检查
// @Description 检查服务状态
// @Tags 系统
// @Success 200 {object} map[string]string "正常"
// @Router /health [get]
func getHealth(c *gyarn.Context) {
	c.JSON(200, map[string]string{"status": "ok"})
}

// getUsers 获取用户
// @Summary 获取用户列表
// @Tags 用户
// @Success 200 {array} User "用户列表"
// @Router /users [get]
func getUsers(c *gyarn.Context) {
	users := []User{{ID: 1, Name: "张三"}, {ID: 2, Name: "李四"}}
	c.JSON(200, users)
}

func main() {
	e := engine.New()

	// 启用OpenAPI - 一行代码！
	docs := openapi.EnableOpenAPI(e, openapi.OpenAPIConfig{
		Title:   "我的API",
		Version: "1.0.0",
	})

	// 从注解生成文档 - 关键的一行！
	docs.GenerateFromAnnotations("./")

	// 注册路由
	e.GET("/health", getHealth)
	e.GET("/users", getUsers)

	log.Println("文档地址: http://localhost:8083/swagger")
	e.Run(":8083")
}
