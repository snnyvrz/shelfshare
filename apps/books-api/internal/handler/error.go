package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/snnyvrz/shelfshare/apps/books-api/internal/validation"
)

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, validation.ErrorResponse{
		Code:    code,
		Message: message,
		Errors:  nil,
	})
}
