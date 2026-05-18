package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Env string
	Port string
}

type DatabaseConfig struct {
	Host string
	Port string
	User string
	Password string
	Name string
	SSLMode string
	URL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: no .env file found, reading from environment")
	}

	cfg := &Config{
		App: AppConfig{
			Env: getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "3001"),
		},
		Database: DatabaseConfig{
			Host: getEnvRequired("DB_HOST"),
			Port: getEnv("DB_PORT", "5432"),
			User: getEnvRequired("DB_USER"),
			Password: getEnvRequired("DB_PASSWORD"),
			Name: getEnvRequired("DB_NAME"),
			SSLMode: getEnv("DB_SSLMODE", "disable"),
		},
	}

	cfg.Database.URL = buildDSN(cfg.Database)

	return cfg
}

func buildDSN(dbConfig DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbConfig.User, 
		dbConfig.Password, 
		dbConfig.Host, 
		dbConfig.Port, 
		dbConfig.Name, 
		dbConfig.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func getEnvRequired(key string) string {
	val, ok := os.LookupEnv(key);

	if !ok || val == "" {
		log.Fatalf("error: required environment variable %q is not set", key)
	}

	return val
}