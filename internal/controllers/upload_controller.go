package controllers

import (
	"net/http"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// UploadConfig contains the upload controller config.
type UploadConfig struct {
	Topic string
}

// UploadController is the controller for the upload route methods.
type UploadController struct {
	config UploadConfig
}

// NewUploadController returns a new UploadService instance.
func NewUploadController(cfg UploadConfig) *UploadController {
	return &UploadController{config: cfg}
}

// Upload godoc
// @Summary File upload
// @Schemes
// @Description Saves a file in the storage service
// @Tags Upload
// @Accept json
// @Success 204
// @Router /upload [post]
func (u UploadController) Upload(c *gin.Context) {
	id := uuid.NewV4().String()
	postView := views.Upload{
		Id: id,
	}

	if err := c.ShouldBindJSON(&postView); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	go func(id string, postView views.Upload) {
		//u.producer.Subscribe(i.topic)

		// logger.Info("Delivering message to topic",
		// 	zap.String("topic", requestID),
		// 	zap.String("requestID", requestID),
		// )

		//err := u.producer.Send(requestID, payload)
		// if err != nil {
		// 	utils.Logger.Error("Produce failed.",
		// 		zap.Error(err.Error()),
		// 		zap.String("requestID", requestID),
		// 	)
		// 	return
		// }
	}(id, postView)

	c.Status(http.StatusNoContent)
}
