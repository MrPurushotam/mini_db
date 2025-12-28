package config

import (
	"os"
)

type Config struct {
	Port     string
	LogLevel string
	AOF_FILENAME string
}

func LoadConfig() *Config {
	port := getEnv("PORT", "3000")
	logLevel := getEnv("LOG_LEVEL", "info")
	filename := getEnv("AOF_FILENAME", "database.aof")

	return &Config{
		Port:     port,
		LogLevel: logLevel,
		AOF_FILENAME: filename,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
