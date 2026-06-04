package handler

import (
	"errors"
	"net/http"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/middleware"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/repository"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var req model.ListTodosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid query params")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid query params",
		})
		return
	}

	response, err := h.service.List(c.Request.Context(), userID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list todos")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("user_id", userID.String()).Msg("Todos of the user with id as \"user_id\"")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "all todos",
		"data": response,
	})
}

func (h *TodoHandler) Create(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var newTodo model.CreateTodoRequest
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		h.log.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid request body",
		})
		return
	}

	todo, err := h.service.Create(c.Request.Context(), userID, &newTodo)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to create todo")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("todo_id", todo.ID.String()).Msg("New todo created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "todo created successfully",
		"data": gin.H{
			"todo": todo,
		},
	})
}

func (h *TodoHandler) Get(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	id := c.Param("id")
	parsedID, err := uuid.Parse(id)

	if err != nil {
		h.log.Warn().Err(err).Msg("invalid id")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid id",
		})
		return
	}

	todo, err := h.service.Get(c.Request.Context(), userID, parsedID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": "todo not found",
			})
			return
		}
		h.log.Error().Err(err).Msg("internal server error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("todo_id", todo.ID.String()).Msg("Todo with id as \"todo_id\"")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "todo fetched successfully",
		"data": gin.H{
			"todo": todo,
		},
	})
}

func (h *TodoHandler) Update(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	id := c.Param("id")
	parsedID, err := uuid.Parse(id)

	if err != nil {
		h.log.Warn().Err(err).Msg("invalid id")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid id",
		})
		return
	}

	var req model.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid request body",
		})
		return
	}

	todo, err := h.service.Update(c.Request.Context(), userID, parsedID, &req)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": "todo not found",
			})
			return
		}
		h.log.Error().Err(err).Msg("internal server error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("todo_id", todo.ID.String()).Msg("Updated todo with id as \"todo_id\"")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "todo updated successfully",
		"data": gin.H{
			"todo": todo,
		},
	})
}



func (h *TodoHandler) Delete(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	id := c.Param("id")
	parsedID, err := uuid.Parse(id)

	if err != nil {
		h.log.Warn().Err(err).Msg("invalid id")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid id",
		})
		return
	}

	deletedID, err := h.service.Delete(c.Request.Context(), userID, parsedID)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": "todo not found",
			})
			return
		}
		h.log.Error().Err(err).Msg("internal server error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("todo_id", deletedID.String()).Msg("Deleted todo with id as \"todo_id\"")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "todo deleted successfully",
		"data": gin.H{
			"id": deletedID,
		},
	})
}