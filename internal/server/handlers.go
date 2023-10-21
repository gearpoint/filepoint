package server

import (
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/controllers"
	"github.com/gin-gonic/gin"

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
	v1.GET("/health", controllers.HealthController{}.HealthCheck)

	upload := v1.Group(string(config.Upload))
	{
		webhookURL := s.routes[config.Upload].WebhookURL
		uploadController := controllers.NewUploadController(
			&controllers.UploadConfig{
				Topic:           s.routes[config.Upload].Topic,
				WebhookURL:      webhookURL,
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
