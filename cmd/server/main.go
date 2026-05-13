package main

import (
	"log"

	"github.com/asad-mujumder/golang-todos-app-rest-api/config"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/router"
)

func main() {
	cfg := config.Load()
	r := router.Setup()

	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatal(err)
	}
}