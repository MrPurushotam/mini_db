package config

import (
	"os"
)

type Config struct {
	Port                     string
	LogLevel                 string
	AOF_FILENAME             string
	AWS_LAMBDA_FUNCTION_NAME string
}

func LoadConfig() *Config {
	port := getEnv("PORT", "3000")
	logLevel := getEnv("LOG_LEVEL", "info")
	filename := getEnv("AOF_FILENAME", "database.aof")
	aws_lambda_name := getEnv("AWS_LAMBDA_FUNCTION_NAME", "")

	return &Config{
		Port:                     port,
		LogLevel:                 logLevel,
		AOF_FILENAME:             filename,
		AWS_LAMBDA_FUNCTION_NAME: aws_lambda_name,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
