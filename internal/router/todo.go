package router

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerTodoRoutes(r *gin.Engine, h *handler.TodoHandler) {
	todos := r.Group("/todos")

	{
		todos.GET("", h.List)
		todos.POST("", h.Create)
		todos.GET("/:id", h.Get)
	}
}