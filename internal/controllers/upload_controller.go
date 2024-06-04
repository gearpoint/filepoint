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
	"github.com/gearpoint/filepoint/config"
	cache_control "github.com/gearpoint/filepoint/internal/cache-control"
	"github.com/gearpoint/filepoint/internal/sender_handlers"
	"github.com/gearpoint/filepoint/internal/uploader"
	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	http_utils "github.com/gearpoint/filepoint/pkg/http"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// The field that contains the file.
	ContentField = "content"
)

// UploadConfig contains the upload controller config.
type UploadConfig struct {
	RouteConfig     config.RouteConfig
	PartitionKey    string
	Publisher       message.Publisher
	AWSRepository   *aws_repository.AWSRepository
	RedisRepository *redis.RedisRepository
}

// UploadController is the controller for the upload route methods.
type UploadController struct {
	tableName     string
	topic         string
	webhookURL    string
	partitionKey  string
	publisher     message.Publisher
	awsRepository *aws_repository.AWSRepository
	cacheControl  *cache_control.UploadCacheControl
}

// NewUploadController returns a new UploadService instance.
func NewUploadController(cfg *UploadConfig) *UploadController {
	return &UploadController{
		tableName:     cfg.RouteConfig.TableName,
		topic:         cfg.RouteConfig.Topic,
		webhookURL:    cfg.RouteConfig.WebhookURL,
		partitionKey:  cfg.PartitionKey,
		publisher:     cfg.Publisher,
		awsRepository: cfg.AWSRepository,
		cacheControl:  cache_control.NewUploadCacheControl(cfg.RedisRepository),
	}
}

// todo: batch upload

// Upload godoc
// @Summary File upload
// @Description Saves a file in the storage service and sends webhook.
// @Tags Upload
// @Accept multipart/form-data
// @Param userId formData string true "User Identifier"
// @Param author formData string false "File upload author"
// @Param title formData string false "File title"
// @Param content formData file true "File to be uploaded"
// @Produce json
// @Success 202
// @Header 202 {object} Webhook-Request-Body "views.WebhookPayload{Id:"X-Request-Id", Success:true, CorrelationId:"", Location:"{location}", Error:""}"
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

	eventType, uploader, err := uploader.GetUploaderByContentType(contentType)
	if err != nil {
		abortWithBadRequest(c, "error validating file content type", err.Error())
		return
	}

	uploadPubSub := &views.UploadPubSub{
		Id:            http_utils.GetRequestId(c),
		UserId:        requestBody.UserId,
		Author:        requestBody.Author,
		Title:         requestBody.Title,
		CorrelationId: requestBody.CorrelationId,
		Filename:      fileHeader.Filename,
		ContentType:   contentType,
		Size:          fileHeader.Size,
		IpAddress:     http_utils.GetIPAddress(c),
		OccurredOn:    time.Now(),
	}

	dynamoDBSchema := views.DynamoDBUploadSchema{
		UserId:        uploadPubSub.UserId,
		Prefix:        utils.GetUniquePrefix(uploadPubSub.UserId),
		Author:        requestBody.Author,
		Title:         requestBody.Title,
		RequestId:     uploadPubSub.Id,
		CorrelationId: uploadPubSub.CorrelationId,
		OccurredOn:    time.Now(),
	}

	uploader.SetConfig(&strategies.UploaderConfig{
		UploadView:    uploadPubSub,
		AWSRepository: u.awsRepository,
		Prefix:        dynamoDBSchema.Prefix,
	})

	err = uploader.Validate(uploadPubSub)
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

	err = u.awsRepository.AddTableRow(u.tableName, dynamoDBSchema)
	if err != nil {
		abortWithBadRequest(c, "error saving file information", err.Error())
		return
	}

	go u.uploadWorker(eventType, uploader, file)

	// Returns the schema of the webhook content.
	c.Header("Webhook-Request-Body", fmt.Sprintf("%#v", views.WebhookPayload{
		Id:            "X-Request-Id",
		Success:       true,
		CorrelationId: "",
		Location:      "{location}",
		Error:         "",
	}))
	c.Status(http.StatusAccepted)
}

// uploadWorker makes the upload publish.
func (u *UploadController) uploadWorker(eventType strategies.EventTypeKey, uploader strategies.Uploader, file multipart.File) {
	cfg := uploader.Config()
	ctx := logger.NewContext(context.Background(), zap.String("request_id", cfg.UploadView.Id))
	logger := logger.WithContext(ctx)

	tempObjectPrefix, err := uploader.UploadTemp(file)
	file.Close()

	if err != nil {
		logger.Error("error saving temp file", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView, u.webhookURL)
		return
	}

	payload, err := json.Marshal(cfg.UploadView)
	if err != nil {
		logger.Error("cannot marshal message", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView, u.webhookURL)
		return
	}

	message := message.NewMessage(cfg.UploadView.Id, payload)
	message.Metadata.Set(views.EventType, string(eventType))
	message.Metadata.Set(views.S3Prefix, cfg.Prefix)
	message.Metadata.Set(views.TempObjectPrefix, tempObjectPrefix)

	if u.partitionKey != "" {
		message.Metadata.Set(u.partitionKey, cfg.UploadView.UserId)
	}

	logger.Info("publishing message to topic", zap.String("topic", u.topic))
	err = u.publisher.Publish(u.topic, message)
	if err != nil {
		logger.Error("error publishing message", zap.Error(err))
		sender_handlers.SendUploadErrorWebhook(ctx, cfg.UploadView, u.webhookURL)
		return
	}
}

// Upload godoc
// @Summary Get file URL
// @Description Returns the file signed URL
// @Tags Upload
// @Param prefix query string true "File folder prefix"
// @Param definition query utils.FileDefinitions false "File definition config"
// @Produce json
// @Success 200 {object} views.GetSignedURLResponse
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [get]
func (u *UploadController) GetSignedURL(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	userId, depth := utils.GetPrefixFolder(prefix)
	if depth != 1 {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	schema := &views.DynamoDBUploadSchema{
		UserId: userId,
		Prefix: prefix,
	}

	err := u.awsRepository.GetTableRow(u.tableName, schema)
	if err != nil {
		logger.Error("error retrieving prefix info from DB",
			zap.Any("prefix", prefix),
			zap.Error(err),
		)
		abortWithBadRequest(c, "error retrieving prefix info")
		return
	}

	definition := utils.AtoFileDefinitions(c.Request.URL.Query().Get("definition"))
	completePrefix := utils.GetClosestPrefix(schema.DefinitionsMap, definition)

	cached, err := u.cacheControl.SignedURLCacheControl.GetBytes(c, completePrefix)
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

	u.cacheControl.SignedURLCacheControl.AddBytes(c, completePrefix, jsonResponse)

	if response.Temporary {
		abortWithBadRequest(c, "temporary file")
		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, jsonResponse)
}

// Upload godoc
// @Summary List files URLs from a folder
// @Description Returns the files signed URLs
// @Tags Upload
// @Param prefix query string true "Folder prefix"
// @Produce json
// @Success 200 {object} []views.ListSignedURLResponse
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload/folder [get]
func (u *UploadController) ListFolder(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the folder prefix is required", "you must provide a valid folder prefix")
		return
	}

	prefixes := u.getFolderPrefixesFullDepth(c, prefix)
	response := u.listPrefixes(c, prefixes)

	c.JSON(http.StatusOK, response)
}

// Upload godoc
// @Summary List files URLs
// @Description Returns the files signed URLs
// @Tags Upload
// @Accept json
// @Param ListObjectsRequest body views.ListObjectsRequest true "List files URLs request body"
// @Produce json
// @Success 200 {object} []views.ListSignedURLResponse
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload/list [post]
func (u *UploadController) ListObjects(c *gin.Context) {
	request := &views.ListObjectsRequest{}
	err := http_utils.ReadRequest(c, request)
	if err != nil || request.Prefixes == nil {
		abortWithBadRequest(c, "the prefixes are required", "you must provide valid prefixes and a valid file definition")
		return
	}

	var completePrefixes []string

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, prefix := range request.Prefixes {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()
			userId, depth := utils.GetPrefixFolder(prefix)
			if depth != 1 {
				logger.Error("unable to retrieve userId",
					zap.String("prefix", prefix),
				)
				return
			}

			schema := &views.DynamoDBUploadSchema{
				UserId: userId,
				Prefix: prefix,
			}

			err := u.awsRepository.GetTableRow(u.tableName, schema)
			if err != nil {
				logger.Error("error retrieving prefix info from DB",
					zap.Any("prefix", prefix),
					zap.Error(err),
				)
				return
			}

			completePrefix := utils.GetClosestPrefix(schema.DefinitionsMap, request.Definition)

			mu.Lock()
			completePrefixes = append(completePrefixes, completePrefix)
			mu.Unlock()
		}(prefix)
	}

	wg.Wait()

	response := u.listPrefixes(c, completePrefixes)

	c.JSON(http.StatusOK, response)
}

// getFolderPrefixesFullDepth returns all prefixes from the given folder.
// It doesn't return folders in the prefixes list, only saved objects.
func (u *UploadController) getFolderPrefixesFullDepth(ctx context.Context, folders ...string) []string {
	var mu sync.Mutex
	var wg sync.WaitGroup

	var prefixes []string
	for _, subPrefix := range folders {
		if !utils.CheckPrefixIsFolder(subPrefix) {
			prefixes = append(
				prefixes,
				subPrefix,
			)
			continue
		}

		newPrefixes, err := u.getFolderPrefixes(ctx, subPrefix)
		if err != nil {
			logger.Warn("error listing prefixes from folder",
				zap.String("folder", subPrefix),
			)
		}

		wg.Add(1)
		go func(folders []string) {
			defer wg.Done()
			toAppend := u.getFolderPrefixesFullDepth(
				ctx,
				folders...,
			)

			mu.Lock()
			prefixes = append(
				prefixes,
				toAppend...,
			)
			mu.Unlock()
		}(newPrefixes)
	}

	wg.Wait()

	return prefixes
}

// getFolderPrefixes returns the given folder prefixes.
func (u *UploadController) getFolderPrefixes(ctx context.Context, folderPrefix string) ([]string, error) {
	prefixes, err := u.cacheControl.PrefixesCacheControl.Get(ctx, folderPrefix)
	if err == nil {
		return prefixes, nil
	}

	prefixes, err = u.awsRepository.ListObjects(folderPrefix)
	if err != nil {
		return nil, err
	}

	u.cacheControl.PrefixesCacheControl.Add(ctx, folderPrefix, prefixes)

	return prefixes, nil
}

// listPrefixes list the given prefixes.
func (u *UploadController) listPrefixes(c context.Context, prefixes []string) []*views.ListSignedURLResponse {
	var mu sync.Mutex
	var wg sync.WaitGroup

	response := []*views.ListSignedURLResponse{}
	for _, objKey := range prefixes {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()

			cached, err := u.cacheControl.SignedURLCacheControl.Get(c, prefix)
			if err == nil {
				if cached.Temporary {
					return
				}

				mu.Lock()
				response = append(response, &views.ListSignedURLResponse{
					prefix: cached,
				})
				mu.Unlock()

				return
			}

			signedUrlResponse, err := u.awsRepository.GetSignedObject(prefix)
			if err != nil {
				logger.Error("error getting signed object", zap.Any("prefix", prefix), zap.Error(err))
				return
			}

			u.cacheControl.SignedURLCacheControl.Add(c, prefix, signedUrlResponse)

			if signedUrlResponse.Temporary {
				return
			}

			mu.Lock()
			response = append(response, &views.ListSignedURLResponse{
				prefix: signedUrlResponse,
			})
			mu.Unlock()
		}(objKey)
	}

	wg.Wait()

	return response
}

// Upload godoc
// @Summary Delete file
// @Description Deletes the file
// @Tags Upload
// @Param prefix query string true "File folder prefix"
// @Produce json
// @Success 200 {string} OK
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload [delete]
func (u *UploadController) Delete(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	userId, depth := utils.GetPrefixFolder(prefix)

	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) || depth != 1 {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	schema := &views.DynamoDBUploadSchema{
		UserId: userId,
		Prefix: prefix,
	}
	err := u.awsRepository.GetTableRow(u.tableName, schema)
	if err != nil {
		logger.Error("error retrieving prefix info from DB",
			zap.Any("prefix", prefix),
			zap.Error(err),
		)
		abortWithBadRequest(c, "error retrieving prefix info")
		return
	}

	for _, filePrefix := range schema.DefinitionsMap {
		err = u.awsRepository.DeleteObject(filePrefix)
		if err != nil {
			logger.Error("error deleting object storage",
				zap.Any("prefix", prefix),
				zap.Error(err),
			)
		}
	}

	u.awsRepository.DelTableRow(u.tableName, schema)
	if err != nil {
		logger.Error("error deleting prefix info from DB",
			zap.Any("prefix", prefix),
			zap.Error(err),
		)
	}
	u.cacheControl.RemoveKeyFromCachedPrefixes(c, prefix)

	c.String(http.StatusOK, "OK")
}

// Upload godoc
// @Summary Delete all
// @Description Deletes all files from prefix
// @Tags Upload
// @Param prefix query string true "File folder prefix"
// @Produce json
// @Success 200 {string} OK
// @Failure 400 {object} http_utils.RestError
// @Failure 500
// @Header all {string} X-Request-Id "Request ID (UUID)"
// @Router /upload/all [delete]
func (u *UploadController) DeleteAll(c *gin.Context) {
	prefix := c.Request.URL.Query().Get("prefix")
	if prefix == "" || !utils.CheckPrefixIsFolder(prefix) {
		abortWithBadRequest(c, "the file prefix is required", "you must provide a valid file prefix")
		return
	}

	prefixes := u.getFolderPrefixesFullDepth(c, prefix)
	if len(prefixes) > 0 {
		// the prefix must be a valid userId otherwise will not exclude.
		err := u.awsRepository.DelTablePartition(u.tableName, "userId", prefix, &views.DynamoDBUploadSchema{})
		if err != nil {
			abortWithBadRequest(c, "error deleting files information", err.Error())
			return
		}
		err = u.awsRepository.DeleteMany(prefixes)
		if err != nil {
			abortWithBadRequest(c, "error deleting files", err.Error())
			return
		}

		u.cacheControl.RemoveFolderFromCache(c, prefix, prefixes)
	}

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
