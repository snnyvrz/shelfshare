package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/go-book-crud-gin/internal/config"
	"github.com/snnyvrz/go-book-crud-gin/internal/db"
	"github.com/snnyvrz/go-book-crud-gin/internal/handler"
	"github.com/snnyvrz/go-book-crud-gin/internal/model"
)

const appVersion = "0.1.0"

func main() {
	startTime := time.Now()

	cfg := config.Load()

	gin.SetMode(cfg.GinMode)

	e := gin.Default()

	e.SetTrustedProxies([]string{
		"127.0.0.1",
		"::1",
	})

	database := db.ConnectWithRetry(cfg)

	if err := database.AutoMigrate(&model.Book{}); err != nil {
		panic(err)
	}

	healthHandler := handler.NewHealthHandler(database, startTime, appVersion)
	healthHandler.RegisterRoutes(e)

	bookHandler := handler.NewBookHandler(database)
	bookHandler.RegisterRoutes(e)

	e.Run(":8080")
}
