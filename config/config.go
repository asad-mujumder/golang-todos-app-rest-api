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

func (a AppConfig) Addr() string {
	return ":" + a.Port
}

type DatabaseConfig struct {
	Driver string
	Host string
	Port string
	User string
	Password string
	Name string
	SSLMode string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		d.Driver,
		d.User, 
		d.Password, 
		d.Host, 
		d.Port, 
		d.Name, 
		d.SSLMode,
	)
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: no .env file found, reading from environment")
	}

	return &Config{
		App: AppConfig{
			Env: getEnvWithDefault("APP_ENV", "development"),
			Port: getEnvWithDefault("APP_PORT", "3001"),
		},
		Database: DatabaseConfig{
			Driver: getEnvWithDefault("DB_DRIVER", "postgres"),
			Host: getEnvWithDefaultRequired("DB_HOST"),
			Port: getEnvWithDefault("DB_PORT", "5432"),
			User: getEnvWithDefaultRequired("DB_USER"),
			Password: getEnvWithDefaultRequired("DB_PASSWORD"),
			Name: getEnvWithDefaultRequired("DB_NAME"),
			SSLMode: getEnvWithDefault("DB_SSLMODE", "disable"),
		},
	}
}

func getEnvWithDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func getEnvWithDefaultRequired(key string) string {
	val, ok := os.LookupEnv(key);

	if !ok || val == "" {
		log.Fatalf("error: required environment variable %q is not set", key)
	}

	return val
}