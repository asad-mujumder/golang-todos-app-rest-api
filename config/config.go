package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	Database DatabaseConfig
	JWT JWTConfig
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

type JWTConfig struct {
	AccessTokenSecret string
	AccessTokenTTL time.Duration
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
			Host: getEnvRequired("DB_HOST"),
			Port: getEnvWithDefault("DB_PORT", "5432"),
			User: getEnvRequired("DB_USER"),
			Password: getEnvRequired("DB_PASSWORD"),
			Name: getEnvRequired("DB_NAME"),
			SSLMode: getEnvWithDefault("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			AccessTokenSecret: getEnvRequired("ACCESS_TOKEN_SECRET"),
			AccessTokenTTL: 15 * time.Minute,
		},
	}
}

func getEnvWithDefault(key, fallback string) string {
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