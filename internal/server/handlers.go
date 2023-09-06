package server

import (
	"github.com/gearpoint/filepoint/docs"
	"github.com/gearpoint/filepoint/internal/controllers"
	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	gswagger "github.com/swaggo/gin-swagger"
)

// MapHandlers maps server handlers.
func (s *Server) MapHandlers() error {
	router := s.Engine

	docs.SwaggerInfo.Title = "Filepoint"
	docs.SwaggerInfo.BasePath = "/v1"

	router.GET("/docs/*any", gswagger.WrapHandler(swaggerfiles.Handler))

	v1 := router.Group("v1")
	v1.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Filepoint! Docs are available at /v1/docs",
		})
	})

	v1.GET("/health", controllers.HealthController{}.HealthCheck)

	upload := v1.Group("upload")
	{
		uploadController := controllers.NewUploadController(
			controllers.UploadConfig{
				Topic: "filepoint.upload.queuing",
			},
		)

		upload.POST("", uploadController.Upload)
	}

	return nil
}
