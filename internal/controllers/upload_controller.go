package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	cache_control "github.com/gearpoint/filepoint/internal/cache-control"
	"github.com/gearpoint/filepoint/internal/sender_handlers"
	"github.com/gearpoint/filepoint/internal/uploader"
	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	http_utils "github.com/gearpoint/filepoint/pkg/http"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/gearpoint/filepoint/pkg/watermill"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// The field that contains the file.
	ContentField = "content"
)

// UploadConfig contains the upload controller config.
type UploadConfig struct {
	Topic           string
	Publisher       message.Publisher
	AWSRepository   *aws_repository.AWSRepository
	RedisRepository *redis.RedisRepository
}

// UploadController is the controller for the upload route methods.
type UploadController struct {
	topic         string
	publisher     message.Publisher
	awsRepository *aws_repository.AWSRepository
	cacheControl  *cache_control.UploadCacheControl
}

// NewUploadController returns a new UploadService instance.
func NewUploadController(cfg *UploadConfig) *UploadController {
	return &UploadController{
		topic:         cfg.Topic,
		publisher:     cfg.Publisher,
		awsRepository: cfg.AWSRepository,
		cacheControl:  cache_control.NewUploadCacheControl(cfg.RedisRepository),
	}
}

// Upload godoc
// @Summary File upload
// @Description Saves a file in the storage service
// @Tags Upload
// @Accept multipart/form-data
// @Param userId formData string true "User Identifier"
// @Param author formData string true "File upload author"
// @Param title formData string true "File title"
// @Param content formData file true "File to be uploaded"
// @Produce json
// @Success 202
// @Header 202 {object} Webhook-Request-Body "views.WebhookPayload{Id:"X-Request-Id", Success:true, Location:"{location}", Labels:[]string{}, Error:""}"
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [post]
func (u *UploadController) Upload(c *gin.Context) {
	requestBody := &views.UploadRequest{}
	if err := http_utils.ReadRequest(c, &requestBody); err != nil {
		abortWithBadRequest(c, "error reading request", err.Error())
		return
	}

	fileHeader, err := http_utils.ReadRequestFile(c, ContentField)
	if err != nil {
		abortWithBadRequest(c, "error getting file contents", err.Error())
		return
	}

	contentType, err := utils.GetFileContentType(fileHeader.Header)
	if err != nil {
		abortWithBadRequest(c, "error getting file content type", err.Error())
		return
	}

	eventType, uploaderType, err := uploader.GetTypeByContentType(contentType)
	if err != nil {
		abortWithBadRequest(c, "error validating file content type", err.Error())
		return
	}

	uploadPubSub := &views.UploadPubSub{
		Id:          http_utils.GetRequestId(c),
		UserId:      requestBody.UserId,
		Author:      requestBody.Author,
		Title:       requestBody.Title,
		Filename:    fileHeader.Filename,
		ContentType: contentType,
		Size:        fileHeader.Size,
		IpAddress:   http_utils.GetIPAddress(c),
		OccurredOn:  time.Now(),
	}

	uploader := uploader.NewUploader(uploaderType, &types.UploaderTypeConfig{
		UploadView:    uploadPubSub,
		AWSRepository: u.awsRepository,
	})

	err = uploader.UploaderType.Validate(uploadPubSub)
	if err != nil {
		errSlice := utils.FormatValidatorErrors(err)
		if errSlice != nil {
			abortWithBadRequest(c, "error validating data", errSlice...)
			return
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		abortWithBadRequest(c, "error reading file", err.Error())
	}

	go u.uploadWorker(eventType, uploader, file)

	c.Header("Webhook-Request-Body", fmt.Sprintf("%#v", views.WebhookPayload{
		Id:       "X-Request-Id",
		Success:  true,
		Location: "{location}",
		Labels:   []string{},
		Error:    "",
	}))
	c.Status(http.StatusAccepted)
}

// uploadWorker makes the upload publish.
func (u *UploadController) uploadWorker(eventType types.UploaderTypes, uploader *uploader.Uploader, file multipart.File) {
	cfg := uploader.UploaderType.GetConfig()
	ctx := logger.NewContext(context.Background(), zap.String("request_id", cfg.UploadView.Id))

	s3Prefix, err := uploader.UploaderType.UploadTemp(file)
	file.Close()

	if err != nil {
		logger.WithContext(ctx).Error("error saving temp file", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView.Id)
		return
	}

	payload, err := json.Marshal(cfg.UploadView)
	if err != nil {
		logger.WithContext(ctx).Error("cannot marshal message", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView.Id)
		return
	}

	message := message.NewMessage(cfg.UploadView.Id, payload)
	message.Metadata.Set(views.EventType, string(eventType))
	message.Metadata.Set(views.S3Prefix, s3Prefix)
	message.Metadata.Set(watermill.KafkaKey, cfg.UploadView.UserId)

	err = u.publisher.Publish(u.topic, message)
	if err != nil {
		logger.WithContext(ctx).Error("error publishing message", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView.Id)
		return
	}
}

// Upload godoc
// @Summary Get file URL
// @Description Returns the file signed URL
// @Tags Upload
// @Param prefix query string true "File prefix"
// @Produce json
// @Success 200 {object} views.GetSignedURLResponse
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [get]
func (u *UploadController) GetSignedURL(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	cached, err := u.cacheControl.SignedURLCacheControl.GetBytes(c, prefix)
	if err == nil {
		c.Data(http.StatusOK, gin.MIMEJSON, cached)
		return
	}

	response, err := u.awsRepository.GetSignedObject(prefix)
	if err != nil {
		if aws_repository.CheckIsNotFoundError(err) {
			abortWithNotFound(c, "prefix not found")
			return
		}

		abortWithBadRequest(c, "error getting signed URL")
		return
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		abortWithBadRequest(c, "error getting signed URL")
		return
	}

	u.cacheControl.SignedURLCacheControl.AddBytes(c, prefix, jsonResponse)

	if response.Temporary {
		abortWithBadRequest(c, "temporary file")
		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, jsonResponse)
}

// Upload godoc
// @Summary List files URL
// @Description Returns the files signed URLs
// @Tags Upload
// @Param prefix query string true "Folder prefix"
// @Produce json
// @Success 200 {object} views.GetSignedURLResponse
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload/list [get]
func (u *UploadController) ListObjects(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	prefixes, err := u.getFolderPrefixes(c, prefix)
	if err != nil {
		abortWithBadRequest(c, "error listing prefixes", err.Error())
		return
	}

	response := u.listPrefixes(c, prefixes)

	c.JSON(http.StatusOK, response)
}

// getFolderPrefixes returns the folder prefixes.
func (u *UploadController) getFolderPrefixes(ctx context.Context, prefixesKey string) ([]string, error) {
	prefixes, err := u.cacheControl.PrefixesCacheControl.Get(ctx, prefixesKey)
	if err == nil {
		return prefixes, nil
	}

	prefixes, err = u.awsRepository.ListObjects(prefixesKey)
	if err != nil {
		return nil, err
	}

	u.cacheControl.PrefixesCacheControl.Add(ctx, prefixesKey, prefixes)

	return prefixes, nil
}

// listPrefixes list the given prefixes.
func (u *UploadController) listPrefixes(c context.Context, prefixes []string) []*views.GetSignedURLResponse {
	var mu sync.Mutex
	var wg sync.WaitGroup

	response := []*views.GetSignedURLResponse{}
	for _, objKey := range prefixes {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()

			cached, err := u.cacheControl.SignedURLCacheControl.Get(c, prefix)
			if err == nil {
				if !cached.Temporary {
					mu.Lock()
					response = append(response, cached)
					mu.Unlock()
				}
				return
			}

			signedUrlResponse, err := u.awsRepository.GetSignedObject(prefix)
			if err != nil {
				logger.Error("error getting signed object", zap.Any("prefix", prefix), zap.Error(err))
				return
			}

			u.cacheControl.SignedURLCacheControl.Add(c, prefix, signedUrlResponse)

			if !signedUrlResponse.Temporary {
				mu.Lock()
				response = append(response, signedUrlResponse)
				mu.Unlock()
			}
		}(objKey)
	}

	wg.Wait()

	return response
}

// Upload godoc
// @Summary Delete file
// @Description Deletes the file
// @Tags Upload
// @Param prefix query string true "File prefix"
// @Produce json
// @Success 200 {string} OK
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [delete]
func (u *UploadController) Delete(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	err := u.awsRepository.DeleteObject(prefix)
	if err != nil {
		abortWithBadRequest(c, "error deleting file", err.Error())
		return
	}

	u.cacheControl.RemoveKeyFromCachedPrefixes(c, prefix)

	c.String(http.StatusOK, "OK")
}

// Upload godoc
// @Summary Delete all
// @Description Deletes all files from prefix
// @Tags Upload
// @Param prefix query string true "File prefix"
// @Produce json
// @Success 200 {string} OK
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [delete]
func (u *UploadController) DeleteAll(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	prefixes, err := u.getFolderPrefixes(c, prefix)
	if err != nil {
		abortWithBadRequest(c, "error deleting file", err.Error())
		return
	}

	err = u.awsRepository.DeleteMany(prefixes)
	if err != nil {
		abortWithBadRequest(c, "error deleting files", err.Error())
		return
	}

	u.cacheControl.RemoveFolderFromCache(c, prefix, prefixes)

	c.String(http.StatusOK, "OK")
}

// abortWithBadRequest aborts the request with a bad request error.
func abortWithBadRequest(c *gin.Context, message string, description ...string) {
	fmtErr := http_utils.NewBadRequestError(message, description...)

	c.Error(fmtErr)
	c.AbortWithStatusJSON(fmtErr.Status(), fmtErr)
}

// abortWithNotFound aborts the request with a not found error.
func abortWithNotFound(c *gin.Context, message string, description ...string) {
	fmtErr := http_utils.NewNotFoundError(message, description...)

	c.Error(fmtErr)
	c.AbortWithStatusJSON(fmtErr.Status(), fmtErr)
}
