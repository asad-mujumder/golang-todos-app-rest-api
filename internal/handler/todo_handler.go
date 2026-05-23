package handler

import (
	"net/http"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type TodoHandler struct {
	service *service.TodoService
	log zerolog.Logger
}

func NewTodoHandler(service *service.TodoService, log zerolog.Logger) *TodoHandler {
	return &TodoHandler{
		service: service,
		log: log.With().Str("handler", "todo").Logger(),
	}
}

func (h *TodoHandler) List(c *gin.Context) {
	var req model.ListTodosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid query params")
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error": err.Error(),
		})
		return
	}

	response, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list todos")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("", "").Msg("")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": response,
	})
}

func (h *TodoHandler) Create(c *gin.Context) {
	var newTodo model.CreateTodoRequest
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		h.log.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error": err.Error(),
		})
		return
	}

	todo, err := h.service.Create(c.Request.Context(), &newTodo)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to create todo")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("todo_id", todo.ID.String()).Msg("New todo created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"status": true,
		"data": gin.H{
			"todo": todo,
		},
	})
}