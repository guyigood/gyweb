package main

import (
	"github.com/gin-gonic/gin"
	"your_project/internal/middleware"
	"your_project/internal/service"
)

func main() {
	r := gin.Default()

	userService := service.NewUserService()

	r.Use(middleware.AuthMiddleware())

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := userService.GetUser(id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, user)
	})

	r.Run(":8080")
}