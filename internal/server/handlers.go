package server

import "github.com/gearpoint/filepoint/internal/controllers"

// Map Server Handlers
func (s *Server) MapHandlers() error {
	router := s.Engine

	// docs.SwaggerInfo.Title = "Go example REST API"
	// router.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := router.Group("v1")
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
