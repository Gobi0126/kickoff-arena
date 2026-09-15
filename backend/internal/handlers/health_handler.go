package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/avanthika/efootball-backend/internal/service"
)

type HealthHandler struct {
	service service.HealthService
}

func NewHealthHandler(service service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Check(c *gin.Context) {
	dbOK, err := h.service.Check(c.Request.Context())
	if err != nil || !dbOK {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": true})
}
