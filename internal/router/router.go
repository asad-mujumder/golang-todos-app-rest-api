package router

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Todo *handler.TodoHandler
}

func Setup(handlers *Handlers) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health Cheack - Base Path
	router.GET("/", func(ctx *gin.Context){
		ctx.JSON(200, gin.H{
			"success": true,
			"Message": "API is working",
		})
	})

	registerTodoRoutes(router, handlers.Todo)

	return router
}