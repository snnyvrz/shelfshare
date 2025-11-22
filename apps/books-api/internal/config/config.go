package config

import (
	"fmt"
	"os"
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
	cfg := &Config{
		GinMode:   getenv("GIN_MODE", "debug"),
		TZ:        getenv("TZ", "UTC"),
		DBHost:    getenv("DB_HOST", "localhost"),
		DBPort:    getenv("DB_PORT", "5432"),
		DBUser:    getenv("DB_USER", "postgres"),
		DBPass:    getenv("DB_PASS", ""),
		DBName:    getenv("DB_NAME", "postgres"),
		DBSSLMode: os.Getenv("DB_SSLMODE"),
	}

	if cfg.DBSSLMode == "" {
		if cfg.GinMode == "release" {
			cfg.DBSSLMode = "require"
		} else {
			cfg.DBSSLMode = "disable"
		}
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
