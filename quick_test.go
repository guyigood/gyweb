package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
)

func main() {
	// 创建引擎实例
	r := engine.New()

	// 注册简单路由测试
	r.GET("/", func(c *gyarn.Context) {
		fmt.Println("处理根路径请求")
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Hello GyWeb!",
		})
	})

	r.GET("/test", func(c *gyarn.Context) {
		fmt.Println("处理/test路径请求")
		c.JSON(http.StatusOK, gyarn.H{
			"message": "Test endpoint works!",
		})
	})

	// 打印启动信息
	fmt.Println("Server starting on :8085")
	fmt.Println("测试URL:")
	fmt.Println("  http://localhost:8085/")
	fmt.Println("  http://localhost:8085/test")

	// 启动服务器
	log.Fatal(r.Run(":8085"))
}
