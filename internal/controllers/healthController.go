package controllers

import (
	"net/http"

	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gin-gonic/gin"
)

// HealthController is used for health check.
type HealthController struct{}

// HealthCheck returns a 200 OK response.

// @BasePath /v1/health

// HealthCheck godoc
// @Summary health check
// @Schemes
// @Description returns a 200 OK response
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} OK
// @Router /health [get]
func (h HealthController) HealthCheck(c *gin.Context) {
	logger.WithContext(c).Info("Health check")

	c.String(http.StatusOK, "OK")
}
