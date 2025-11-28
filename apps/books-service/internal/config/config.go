package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GinMode   string
	TZ        string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		GinMode:   getenv("GIN_MODE", "debug"),
		TZ:        getenv("TZ", "UTC"),
		DBHost:    getenv("POSTGRES_HOST", "localhost"),
		DBPort:    getenv("POSTGRES_PORT", "5432"),
		DBUser:    getenv("POSTGRES_USER", "postgres"),
		DBPass:    getenv("POSTGRES_PASSWORD", ""),
		DBName:    getenv("POSTGRES_DB", "postgres"),
		DBSSLMode: "disable",
	}

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.DBHost,
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBPort,
		c.DBSSLMode,
		c.TZ,
	)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
