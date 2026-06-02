package main

import (
	"context"

	"github.com/asad-mujumder/golang-todos-app-rest-api/config"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/database/postgres"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/handler"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/repository"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/router"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/service"
	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/jwt"
	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.App.Env)

	ctx := context.Background() // Root Context for the entire application

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()
	log.Info().Msg("postgres connection pool established")

	// jwt managers
	jwtAccessTokenManager := jwt.NewManager(jwt.Config{ Secret: cfg.JWT.AccessTokenSecret, TTL: cfg.JWT.AccessTokenTTL})

	// repositories
	todoRepo := repository.NewTodoRepository(pool, log)
	authRepo := repository.NewAuthRepository(pool, log)

	// services
	todoService := service.NewTodoService(todoRepo, log)
	authService := service.NewAuthService(authRepo, jwtAccessTokenManager, log)

	// handlers
	todoHandler := handler.NewTodoHandler(todoService, log)
	authHandler := handler.NewAuthHandler(authService, log)

	// router
	r := router.Setup(&router.Handlers{
		Auth: authHandler,
		Todo: todoHandler,
	}, jwtAccessTokenManager, log)

	log.Info().Str("port", cfg.App.Port).Msg("starting server")
	if err := r.Run(cfg.App.Addr()); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}