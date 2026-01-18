package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey string
	ServerPort   string
	LogLevel     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	cfg := &Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		ServerPort:   getEnvOrDefault("SERVER_PORT", "8080"),
		LogLevel:     getEnvOrDefault("LOG_LEVEL", "info"),
	}

	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
