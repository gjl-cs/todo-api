package web

import (
	"github.com/gin-gonic/gin"
	"todo/middleware"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "todo/docs"
)

func StartServer() {
	router := gin.Default()
	router.Use(middleware.Logger())
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Todo API启动成功",
		})
	})
	router.GET("/todos", GetTodos)
	router.GET("/todos/:id", GetTodoByID)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.POST("/todos", CreateTodo)
	router.PUT("/todos/:id", UpdateTodo)
	router.PUT("/todos/batch-complete", BatchCompleteTodos)
	router.PUT("/todos/:id/complete", CompleteTodo)
	router.DELETE("/todos/:id", DeleteTodo)
	router.Run(":8080")
}
