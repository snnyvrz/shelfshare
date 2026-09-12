package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

func findRepoRoot() string {
	dir, _ := os.Getwd()

	for {
		candidate := filepath.Join(dir, ".env.dev")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal(".env.dev not found in any parent directory")
		}
		dir = parent
	}
}

func Load() *Config {
	env := getenv("GIN_MODE", "debug")

	if env == "debug" {
		filename := ".env.dev"
		root := findRepoRoot()
		envPath := filepath.Join(root, filename)

		if err := godotenv.Load(envPath); err != nil {
			log.Printf("warning: could not load %s: %v", envPath, err)
		} else {
			log.Printf("loaded %s from %s", filename, envPath)
		}
	}

	cfg := &Config{
		GinMode:   getenv("GIN_MODE", "debug"),
		TZ:        getenv("TZ", "UTC"),
		DBHost:    getenv("DB_HOST", ""),
		DBPort:    getenv("DB_PORT", ""),
		DBUser:    getenv("DB_USER", ""),
		DBPass:    getenv("DB_PASS", ""),
		DBName:    getenv("DB_NAME", ""),
		DBSSLMode: getenv("DB_SSLMODE", "disable"),
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
