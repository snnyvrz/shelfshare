package db

import (
	"log"
	"time"

	"github.com/snnyvrz/shelfshare/apps/books-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxAttempts     = 10
	defaultDelayBetweenTry = 2 * time.Second
)

func ConnectWithRetry(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= defaultMaxAttempts; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
		if err == nil {
			sqlDB, err2 := db.DB()
			if err2 == nil {
				pingErr := sqlDB.Ping()
				if pingErr == nil {
					return db
				}
				err = pingErr
			} else {
				err = err2
			}
		}

		log.Printf("db not ready (attempt %d/%d): %v", attempt, defaultMaxAttempts, err)
		time.Sleep(defaultDelayBetweenTry)
	}

	log.Fatalf("could not connect to db after %d attempts: %v", defaultMaxAttempts, err)
	return nil
}
