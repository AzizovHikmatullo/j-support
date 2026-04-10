package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		Host     string
		Port     string
		DBName   string
		User     string
		Password string
	}
	JWT struct {
		Secret string
	}
	CORS struct {
		AllowedOrigins []string
	}
	WS struct {
		AllowedOrigins []string
	}
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	cfg := &Config{}
	cfg.Server.Port = os.Getenv("SERVER_PORT")

	cfg.Database.Host = os.Getenv("POSTGRES_HOST")
	cfg.Database.Port = os.Getenv("POSTGRES_PORT")
	cfg.Database.DBName = os.Getenv("POSTGRES_DB")
	cfg.Database.User = os.Getenv("POSTGRES_USER")
	cfg.Database.Password = os.Getenv("POSTGRES_PASSWORD")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	cfg.CORS.AllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")

	cfg.WS.AllowedOrigins = strings.Split(os.Getenv("WS_ORIGINS"), ",")

	return cfg, nil
}
