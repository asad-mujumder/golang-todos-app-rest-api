package main

import (
	"github.com/asad-mujumder/golang-todos-app-rest-api/config"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/database/postgres"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/router"
	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.App.Env)

	pool, err := postgres.NewPool(cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()
	log.Info().Msg("postgres connection pool established")

	r := router.Setup()

	log.Info().Str("port", cfg.App.Port).Msg("starting server")
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}