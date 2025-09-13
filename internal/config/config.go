package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBTimezone string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTLSec   int

	ElasticAddr     string
	ElasticUsername string
	ElasticPassword string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvi(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		return n
	}
	return def
}

func Load() *Config {
	return &Config{
		Port: getenv("PORT", "8080"),

		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", "postgres"),
		DBName:     getenv("DB_NAME", "blog"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		DBTimezone: getenv("DB_TIMEZONE", "UTC"),

		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       getenvi("REDIS_DB", 0),
		CacheTTLSec:   getenvi("CACHE_TTL_SECONDS", 300),

		ElasticAddr:     getenv("ELASTICSEARCH_ADDR", "http://localhost:9200"),
		ElasticUsername: getenv("ELASTICSEARCH_USERNAME", ""),
		ElasticPassword: getenv("ELASTICSEARCH_PASSWORD", ""),
	}
} 