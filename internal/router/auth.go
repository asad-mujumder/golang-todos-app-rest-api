package router

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(r *gin.Engine, h *handler.AuthHandler) {
	auth := r.Group("/auth")

	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		// auth.POST("/refresh-token")
		// auth.POST("/forgot-password")
		// auth.POST("/reset-password")
		// auth.POST("/confirm-email")
	}
}