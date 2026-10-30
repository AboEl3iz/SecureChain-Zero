package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration for the microservice.
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ServiceName  string
	Environment  string
	Version      string
}

// LoadFromEnv loads configuration from environment variables with safe defaults.
func LoadFromEnv() Config {
	port := 8080
	if val := os.Getenv("PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			port = p
		}
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	return Config{
		Port:         port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ServiceName:  "secure-app",
		Environment:  env,
		Version:      version,
	}
}
