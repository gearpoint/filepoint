package middlewares

import (
	"strconv"
	"time"

	"github.com/gearpoint/filepoint/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// MetricsMiddleware collects HTTP metrics for Prometheus
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get request size
		if c.Request.ContentLength > 0 {
			metrics.Get().HTTPRequestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		// Record metrics
		metrics.Get().HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			status,
		).Inc()

		metrics.Get().HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			endpoint,
			status,
		).Observe(duration)

		// Record response size
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			metrics.Get().HTTPResponseSize.WithLabelValues(
				c.Request.Method,
				endpoint,
			).Observe(float64(responseSize))
		}
	}
}
