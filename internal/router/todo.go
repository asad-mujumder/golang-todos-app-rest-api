package router

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerTodoRoutes(r *gin.Engine, h *handler.TodoHandler) {
	todos := r.Group("/todos")

	{
		todos.POST("", h.Create)
	}
}