package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	FrontendOrigin  string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:4200"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
