package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db        *gorm.DB
	startTime time.Time
	version   string
}

func NewHealthHandler(db *gorm.DB, startTime time.Time, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: startTime,
		version:   version,
	}
}

func (h *HealthHandler) RegisterRoutes(e *gin.Engine) {
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)
}

func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": h.version,
		"uptime":  int64(uptime.Seconds()),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "failed to get underlying DB",
		})
		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"db": gin.H{
				"status": "down",
				"error":  err.Error(),
			},
		})
		return
	}

	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"version": h.version,
		"uptime":  int64(uptime.Seconds()),
		"db": gin.H{
			"status": "up",
		},
	})
}
