package main

import (
	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/go-book-crud-gin/internal/handler"
)

func main() {
	e := gin.Default()

	e.SetTrustedProxies([]string{
		"127.0.0.1",
		"::1",
	})

	healthHandler := handler.NewHealthHandler()
	healthHandler.RegisterRoutes(e)

	e.Run(":8080")
}
