package config

import "os"

type ConfigVar struct {
	DB_URL               string
	PORT                 string
	REDIS_URL            string
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string
	GOOGLE_REDIRECT_URL  string
	FRONTEND_URL         string
	JWT_SECRET           string
	RESEND_KEY           string
	GEOIP_DB_PATH        string
	COOKIE_DOMAIN        string
	COOKIE_SECURE        bool
}

func LoadConfig() *ConfigVar {
	return &ConfigVar{
		DB_URL:               getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/biolynq_db?sslmode=disable"),
		PORT:                 getEnv("PORT", "8000"),
		REDIS_URL:            getEnv("REDIS_URL", "localhost:6379"),
		GOOGLE_CLIENT_ID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GOOGLE_REDIRECT_URL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8000/api/v1/auth/google/callback"),
		FRONTEND_URL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		JWT_SECRET:           getEnv("JWT_SECRET", "super-secret-biolynq-auth-jwt-token-key-change-in-production"),
		RESEND_KEY:           os.Getenv("RESEND_KEY"),
		GEOIP_DB_PATH:        getEnv("GEOIP_DB_PATH", "resources/geoip/GeoLite2-City.mmdb"),
		COOKIE_DOMAIN:        getEnv("COOKIE_DOMAIN", ""),
		COOKIE_SECURE:        os.Getenv("COOKIE_SECURE") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
