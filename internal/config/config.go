package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

// Load читает настройки из переменных окружения.
// Значения по умолчанию совпадают с docker-compose.yml.
func Load() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://dentist:dentist@localhost:5433/dentist?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
