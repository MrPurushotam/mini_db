package config

import (
	"os"
)

type Config struct {
	Port     string
	LogLevel string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	logLevel := getEnv("LOG_LEVEL", "info")

	return &Config{
		Port:     port,
		LogLevel: logLevel,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
