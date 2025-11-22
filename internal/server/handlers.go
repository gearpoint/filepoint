package server

import (
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/controllers"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerfiles "github.com/swaggo/files"
	gswagger "github.com/swaggo/gin-swagger"
)

// MapHandlers maps server handlers.
func (s *Server) MapHandlers() error {
	router := s.Engine

	v1 := router.Group("/v1")
	v1.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Filepoint! Docs are available at /v1/docs/index.html",
		})
	})

	v1.GET("/docs/*any", gswagger.WrapHandler(swaggerfiles.Handler))

	// Metrics endpoint for Prometheus
	v1.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoints
	healthController := controllers.NewHealthController(
		s.awsRepository,
		s.redisRepository,
		logger.NewLogger(s.config.Debug),
	)
	v1.GET("/health", healthController.HealthCheck)
	v1.GET("/health/live", healthController.LivenessCheck)
	v1.GET("/health/ready", healthController.ReadinessCheck)

	upload := v1.Group(string(config.Upload))
	{
		uploadController := controllers.NewUploadController(
			&controllers.UploadConfig{
				RouteConfig:     s.routes[config.Upload],
				PartitionKey:    s.partitionKey,
				Publisher:       s.publisher,
				AWSRepository:   s.awsRepository,
				RedisRepository: s.redisRepository,
			},
		)

		upload.GET("", uploadController.GetSignedURL)
		upload.GET("/folder", uploadController.ListFolder)
		upload.POST("", uploadController.Upload)
		upload.POST("/list", uploadController.ListObjects)
		upload.DELETE("", uploadController.Delete)
		upload.DELETE("/all", uploadController.DeleteAll)
	}

	return nil
}
