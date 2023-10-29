package sender_handlers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	cache_control "github.com/gearpoint/filepoint/internal/cache-control"
	"github.com/gearpoint/filepoint/internal/uploader"
	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/watermill"
	"go.uber.org/zap"
)

const (
	// Filepoint get image route.
	getSignedURLBase = "/v1/upload?prefix="
)

type UploadHandler struct {
	maxRetries         int
	poisonQueueTopic   string
	webhookURL         string
	awsRepository      *aws_repository.AWSRepository
	uploadCacheControl *cache_control.UploadCacheControl
}

func NewUploadHandler(awsRepository *aws_repository.AWSRepository, redisRepository *redis.RedisRepository, webhookURL string) *UploadHandler {
	return &UploadHandler{
		maxRetries:         10,
		poisonQueueTopic:   "upload_poison_queue",
		webhookURL:         webhookURL,
		awsRepository:      awsRepository,
		uploadCacheControl: cache_control.NewUploadCacheControl(redisRepository),
	}
}

// ProccessUploadMessages proccess the upload and returns the callback message.
func (h *UploadHandler) ProccessUploadMessages() message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msg.SetContext(logger.NewContext(msg.Context(), zap.String("request_id", msg.UUID)))
		eventType := types.UploaderTypes(msg.Metadata.Get(views.EventType))

		for key, typeMapping := range uploader.UploaderTypesMap {
			if key == eventType {
				var uploadPubSub = &views.UploadPubSub{}
				err := json.Unmarshal(msg.Payload, uploadPubSub)
				if err != nil {
					logger.WithContext(msg.Context()).Error(err.Error())
					return nil, err
				}

				location, labels, err := h.handleUpload(msg, &typeMapping, uploadPubSub)
				if err != nil {
					logger.WithContext(msg.Context()).Error(err.Error())
					return nil, err
				}

				webhookPayload, err := json.Marshal(views.WebhookPayload{
					Id:       uploadPubSub.Id,
					Success:  err == nil,
					Location: location,
					Labels:   labels,
					Error:    "",
				})
				if err != nil {
					logger.WithContext(msg.Context()).Error(err.Error())
					return nil, err
				}

				msg.Ack()

				return message.Messages{
					message.NewMessage(uploadPubSub.Id, webhookPayload),
				}, nil
			}
		}

		return nil, errors.New("unrecognized event-type")
	}
}

// handleUpload is responsible for uploading the file.
func (h *UploadHandler) handleUpload(msg *message.Message, uploaderType *types.TypeMapping, uploadPubSub *views.UploadPubSub) (string, []string, error) {
	uploader := uploader.NewUploader(uploaderType, &types.UploaderTypeConfig{
		UploadView:    uploadPubSub,
		AWSRepository: h.awsRepository,
	})

	s3Prefix := msg.Metadata.Get(views.S3Prefix)
	reader, err := uploader.UploaderType.HandleFile(s3Prefix)
	if err != nil {
		return "", nil, err
	}

	newS3Prefix, err := uploader.UploaderType.Upload(reader)
	reader.Close()

	if err != nil {
		return "", nil, err
	}

	location := getSignedURLBase + newS3Prefix
	labels := uploader.UploaderType.GetLabels(newS3Prefix)
	labels = append(labels, msg.Metadata.Get(views.EventType))

	tagging := make(map[string]string, len(labels))
	if len(labels) > 0 {
		for _, label := range labels {
			if _, ok := tagging[label]; ok {
				continue
			}
			tagging[label] = ""
		}
	}

	h.awsRepository.PutObjectTagging(newS3Prefix, tagging)
	h.uploadCacheControl.PrefixesCacheControl.AddKeyToCachedPrefixes(msg.Context(), newS3Prefix)

	return location, labels, nil
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
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second * 5,
		Multiplier:      1.2,
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
			logger.Info("sending error message to webhook...")
			SendUploadErrorWebhook(msg.Context(), msg.UUID, h.webhookURL)
			msg.Ack()
		}
	}(messages)
}

// SendUploadErrorWebhook calls the upload webhook with error message.
func SendUploadErrorWebhook(ctx context.Context, id string, webhookURL string) {
	httpPublisher, err := watermill.NewHttpPublisher()
	if err != nil {
		logger.WithContext(ctx).Error("error initializing http publisher", zap.Error(err))
		return
	}

	payload, err := json.Marshal(views.WebhookPayload{
		Id:       id,
		Success:  false,
		Location: "",
		Labels:   []string{},
		Error:    "error uploading file",
	})

	if err != nil {
		logger.WithContext(ctx).Error("cannot marshal message", zap.Error(err))
		return
	}

	message := message.NewMessage(id, payload)

	err = httpPublisher.Publish(webhookURL, message)
	if err != nil {
		logger.WithContext(ctx).Error("error sending http request", zap.Error(err))
		return
	}

	logger.Info("sending error message to webhook...")
}
