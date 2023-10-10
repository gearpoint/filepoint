package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController is used for health check.
type HealthController struct{}

// HealthCheck godoc
// @Summary Health check
// @Description Returns a 200 OK response
// @Tags HealthCheck
// @Produce plain
// @Success 200 {string} OK
// @Router /health [get]
func (h HealthController) HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
