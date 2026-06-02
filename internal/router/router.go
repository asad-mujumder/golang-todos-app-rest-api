package router

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/middleware"
	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Auth *handler.AuthHandler
	Todo *handler.TodoHandler
}

func Setup(handlers *Handlers,  jwtManager *jwt.Manager, log zerolog.Logger) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health Cheack - Base Path
	router.GET("/", func(ctx *gin.Context){
		ctx.JSON(200, gin.H{
			"success": true,
			"Message": "API is working",
		})
	})

	registerAuthRoutes(router, handlers.Auth)

	// protected routes
	protected := router.Group("", middleware.Auth(jwtManager, log))
	registerTodoRoutes(protected, handlers.Todo)

	return router
}