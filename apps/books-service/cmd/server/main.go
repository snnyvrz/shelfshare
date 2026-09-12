package main

// @title           Shelfshare Books API
// @version         1.0
// @description     API for managing books in Shelfshare.

// @contact.name   Sina Niyavarzi
// @contact.email  sinaniya@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/config"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/db"
	docs "github.com/snnyvrz/shelfshare/apps/books-service/internal/docs"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/handler"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/model"
	"github.com/snnyvrz/shelfshare/apps/books-service/internal/repository"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	docs.SwaggerInfo.BasePath = "/api"

	database := db.ConnectWithRetry(cfg)

	if err := database.AutoMigrate(&model.Author{}, &model.Book{}); err != nil {
		panic(err)
	}

	healthHandler := handler.NewHealthHandler(database, startTime, appVersion)
	healthHandler.RegisterRoutes(e)

	api := e.Group("/api")
	{
		bookRepo := repository.NewGormBookRepository(database)
		authorRepo := repository.NewAuthorRepository(database)

		bookHandler := handler.NewBookHandler(bookRepo)
		authorHandler := handler.NewAuthorHandler(authorRepo)

		bookHandler.RegisterRoutes(api)
		authorHandler.RegisterRoutes(api)
	}

	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	e.Run("0.0.0.0:8080")
}
