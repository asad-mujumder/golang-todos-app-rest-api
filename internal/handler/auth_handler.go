package handler

import (
	"errors"
	"net/http"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/repository"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type AuthHandler struct {
	service *service.AuthService
	log zerolog.Logger
}

func NewAuthHandler(service *service.AuthService, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		log: log.With().Str("handler", "auth").Logger(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid request body",
		})
		return
	}

	newUserID, err := h.service.Register(c.Request.Context(), &req)

	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": "email already exists",
			})
			return
		}
		h.log.Error().Err(err).Msg("failed to register user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Info().Str("user_id", *newUserID).Msg("new user created")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "registration successful",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "invalid request body",
		})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": "invalid credentials",
			})
			return
		}

		h.log.Error().Err(err).Msg("failed to log in")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "internal server error",
		})
		return
	}

	h.log.Warn().Str("user_id", resp.User.ID.String()).Msg("user logged in")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged in successfully",
		"data": resp,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}