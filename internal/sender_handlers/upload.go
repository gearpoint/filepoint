package sender_handlers

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gearpoint/filepoint/config"
	cache_control "github.com/gearpoint/filepoint/internal/cache-control"
	"github.com/gearpoint/filepoint/internal/uploader"
	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/gearpoint/filepoint/pkg/watermill"
	"go.uber.org/zap"
)

type UploadHandler struct {
	tableName          string
	maxRetries         int
	poisonQueueTopic   string
	webhookURL         string
	awsRepository      *aws_repository.AWSRepository
	uploadCacheControl *cache_control.UploadCacheControl
}

func NewUploadHandler(awsRepository *aws_repository.AWSRepository, redisRepository *redis.RedisRepository, routeCfg config.RouteConfig) *UploadHandler {
	return &UploadHandler{
		tableName:          routeCfg.TableName,
		maxRetries:         routeCfg.MaxRetries,
		poisonQueueTopic:   routeCfg.PoisonTopic,
		webhookURL:         routeCfg.WebhookURL,
		awsRepository:      awsRepository,
		uploadCacheControl: cache_control.NewUploadCacheControl(redisRepository),
	}
}

func SetMessageContext(msg *message.Message) {
	msg.SetContext(
		logger.NewContext(
			msg.Context(),
			zap.String("request_id", msg.UUID),
		),
	)
}

// ProccessUploadMessages proccess the upload and returns the callback message.
func (h *UploadHandler) ProccessUploadMessages() message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msg.SetContext(
			logger.NewContext(
				msg.Context(),
				zap.String("request_id", msg.UUID),
			),
		)

		logger := logger.WithContext(msg.Context())

		eventType := strategies.EventTypeKey(msg.Metadata.Get(views.EventType))
		s3Prefix := msg.Metadata.Get(views.S3Prefix)

		logger.Info("processing upload message",
			zap.Any("eventType", eventType),
			zap.Any("s3Prefix", s3Prefix),
		)

		uploadPubSub, err := h.unmarshalUpload(msg.Payload)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		err = h.handleUpload(msg, uploadPubSub)
		if err != nil {
			return nil, err
		}

		logger.Info("sending success upload webhook")

		webhookPayload, err := json.Marshal(views.WebhookPayload{
			Id:            uploadPubSub.Id,
			Success:       err == nil,
			CorrelationId: uploadPubSub.CorrelationId,
			Location:      s3Prefix,
			Error:         "",
		})
		if err != nil {
			return nil, err
		}

		h.uploadCacheControl.PrefixesCacheControl.AddKeyToCachedPrefixes(msg.Context(), s3Prefix)

		msg.Ack()

		return message.Messages{
			message.NewMessage(uploadPubSub.Id, webhookPayload),
		}, nil
	}
}

// unmarshalUpload returns the UploadPubSub view or error.
func (h *UploadHandler) unmarshalUpload(payload message.Payload) (*views.UploadPubSub, error) {
	var uploadPubSub = &views.UploadPubSub{}
	err := json.Unmarshal(payload, uploadPubSub)

	return uploadPubSub, err
}

// handleUpload is responsible for uploading the file.
func (h *UploadHandler) handleUpload(msg *message.Message, uploadPubSub *views.UploadPubSub) error {
	eventType := strategies.EventTypeKey(msg.Metadata.Get(views.EventType))
	s3Prefix := msg.Metadata.Get(views.S3Prefix)
	tempObjectPrefix := msg.Metadata.Get(views.TempObjectPrefix)

	ctx := logger.NewContext(msg.Context(), zap.Any("s3Prefix", s3Prefix))
	logger := logger.WithContext(ctx)

	schema := &views.DynamoDBUploadSchema{
		UserId: uploadPubSub.UserId,
		Prefix: s3Prefix,
	}
	err := h.awsRepository.GetTableRow(
		h.tableName, schema,
	)
	if err != nil {
		logger.Error("error retrieving table info from DB",
			zap.String("tableName", h.tableName),
			zap.Error(err),
		)
		return errors.New("error retrieving table info from DB")
	}

	uploader, err := uploader.GetUploaderByEventType(eventType)
	if err != nil {
		logger.Error("unrecognized event-type",
			zap.Error(err),
		)
		return errors.New("unrecognized event-type")
	}

	uploader.SetConfig(&strategies.UploaderConfig{
		UploadView:    uploadPubSub,
		AWSRepository: h.awsRepository,
		Prefix:        s3Prefix,
	})

	tempReader, err := uploader.DownloadTemp(tempObjectPrefix)
	if err != nil {
		logger.Error("error downloading temp file",
			zap.Error(err),
		)
		return err
	}
	filename, err := utils.CreateTmpFile(tempReader)
	tempReader.Close()
	if err != nil {
		logger.Error("error creating temp file",
			zap.Error(err),
		)
		return err
	}
	defer os.Remove(filename)

	fileDefs := uploader.FileDefinitions()
	definitionsMap := utils.FileDefinitionsMapping{}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for def, name := range fileDefs {
		wg.Add(1)
		go func(
			def utils.FileDefinitions, name string, uploader strategies.Uploader, logger *zap.Logger,
		) {
			defer wg.Done()
			aUploader := uploader
			aLogger := logger

			handledReader, err := aUploader.HandleFile(def, filename)
			if err != nil {
				aLogger.Warn("error handling the file",
					zap.Any("definition", def),
					zap.Error(err),
				)
				return
			}
			defer handledReader.Close()

			objectName, err := aUploader.Upload(name, handledReader)
			if err != nil {
				aLogger.Warn("error uploading file to storage",
					zap.Any("definition", def),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			definitionsMap[def] = objectName
			mu.Unlock()

			logger.Info("file uploaded successfully",
				zap.String("objectName", objectName),
			)
		}(def, name, uploader, logger)
	}
	wg.Wait()

	if len(fileDefs) == 0 {
		return errors.New("file could not be uploaded")
	}

	schema.DefinitionsMap = definitionsMap

	err = h.awsRepository.UpdateTableRow(
		h.tableName, schema,
	)
	if err != nil {
		logger.Error("error updating table row",
			zap.Any("tableName", h.tableName),
			zap.Any("userId", uploadPubSub.UserId),
			zap.Error(err),
		)
		return errors.New("unable to update file data in DB")
	}

	return nil
}

// SetupUploadMiddlewares returns the specific upload middlewares.
func (h *UploadHandler) SetupUploadMiddlewares() []message.HandlerMiddleware {
	gochannel := watermill.NewGoChannel()

	poisonQueue, err := middleware.PoisonQueue(gochannel, h.poisonQueueTopic)
	if err != nil {
		panic(err)
	}
	go h.processUploadPoisonQueue(gochannel, h.poisonQueueTopic)

	retryMiddleware := middleware.Retry{
		MaxRetries:      h.maxRetries,
		InitialInterval: time.Second * 5,
		MaxInterval:     time.Hour * 5,
		Multiplier:      1.25,
		Logger:          watermill.NewZapLoggerAdapter(logger.Logger),
	}

	return []message.HandlerMiddleware{
		poisonQueue,
		retryMiddleware.Middleware,
	}
}

// processUploadPoisonQueue consumes the messages coming from poison queue.
func (h *UploadHandler) processUploadPoisonQueue(gochannel *gochannel.GoChannel, poisonQueueTopic string) {
	messages, err := gochannel.Subscribe(context.Background(), poisonQueueTopic)
	if err != nil {
		logger.Error("unable to publish error messages")
	}

	go func(messages <-chan *message.Message) {
		for msg := range messages {
			uploadPubSub, err := h.unmarshalUpload(msg.Payload)
			if err != nil {
				uploadPubSub = &views.UploadPubSub{
					Id: msg.UUID,
				}
			}

			logger.Info("sending error message to webhook...")
			SendUploadErrorWebhook(msg.Context(), uploadPubSub, h.webhookURL)
			msg.Ack()
		}
	}(messages)
}

// SendUploadErrorWebhook calls the upload webhook with error message.
func SendUploadErrorWebhook(ctx context.Context, uploadPubSub *views.UploadPubSub, webhookURL string) {
	logger := logger.WithContext(ctx)

	httpPublisher, err := watermill.NewHttpPublisher()
	if err != nil {
		logger.Error("error initializing http publisher", zap.Error(err))
		return
	}

	logger.Info("sending error upload webhook")

	payload, err := json.Marshal(views.WebhookPayload{
		Id:            uploadPubSub.Id,
		Success:       false,
		CorrelationId: uploadPubSub.CorrelationId,
		Location:      "",
		Error:         "error uploading file",
	})

	if err != nil {
		logger.Error("cannot marshal message", zap.Error(err))
		return
	}

	message := message.NewMessage(uploadPubSub.Id, payload)

	err = httpPublisher.Publish(webhookURL, message)
	if err != nil {
		logger.Error("error sending http request", zap.Error(err))
		return
	}

	logger.Info("sending error message to webhook...")
}
