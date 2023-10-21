// Filepoint is the Gearpoint's file manager service. It's built for performance.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	config "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/sender_handlers"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/gearpoint/filepoint/pkg/watermill"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	// configFile is the main configuration filepath.
	configFile string
)

func main() {
	godotenv.Load()

	envType := utils.GetEnvironmentType()
	switch envType {
	case utils.Development:
		logger.InitLogger(logger.DevelopmentMode)
	case utils.Production:
		logger.InitLogger(logger.ProductionMode)
	default:
		log.Fatal("error initializing logger")
	}

	logger.Info("starting Filepoint server...")

	flag.StringVar(&configFile, "config", "./config/config-local.yaml", "aaa")
	flag.Parse()

	viperConfig, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Fatal("error initializing config",
			zap.Error(err),
		)
	}

	cfg, err := config.ParseConfig(viperConfig)
	if err != nil {
		logger.Fatal("error getting config",
			zap.Error(err),
		)
	}

	setupRouter(cfg)
}

// setupRouter starts the router with the pub/sub configuration.
func setupRouter(cfg *config.Config) {
	context := context.Background()

	awsRepository, err := aws_repository.NewAWSRepository(&cfg.AWSConfig, context)
	if err != nil {
		logger.Fatal("cannot initialize storage client",
			zap.Error(err),
		)
	}

	redisRepository := redis.NewRedisRepository(&cfg.RedisConfig)
	defer redisRepository.Client.Close()
	logger.Info("Redis connected")

	subscriber, err := watermill.NewKafkaSubscriber(&cfg.KafkaConfig)
	if err != nil {
		logger.Fatal("error initializing Kafka subscriber",
			zap.Error(err),
		)
	}

	publisher, err := watermill.NewHttpPublisher()
	if err != nil {
		logger.Fatal("error initializing HTTP publisher",
			zap.Error(err),
		)
	}

	router, err := watermill.NewRouter()
	if err != nil {
		logger.Fatal("error initializing Router",
			zap.Error(err),
		)
	}
	defer router.Close()

	throttleMiddleware := middleware.NewThrottle(
		cfg.KafkaConfig.MessagesPerSecond,
		time.Second,
	)

	router.AddMiddleware(
		middleware.Recoverer,
		throttleMiddleware.Middleware,
	)

	router.AddPlugin(plugin.SignalsHandler)

	for routeName, routeConfig := range cfg.Routes {
		switch routeName {
		case config.Upload:
			webhookURL := cfg.Routes[routeName].WebhookURL
			upload := sender_handlers.NewUploadHandler(awsRepository, redisRepository, webhookURL)
			uploadHandler := router.AddHandler(
				string(routeName),
				routeConfig.Topic,
				subscriber,
				webhookURL,
				publisher,
				upload.ProccessUploadMessages(),
			)
			uploadHandler.AddMiddleware(upload.SetupUploadMiddlewares()...)
		default:
			logger.Warn("no config found for provided route",
				zap.Any("route_name", routeName),
			)
		}
	}

	var version string

	versionByte, err := os.ReadFile("VERSION")
	if err == nil {
		version = string(versionByte)
	}

	logger.Info("Filepoint webhook provider on!",
		zap.String("version", version),
	)

	err = router.Run(context)
	if err != nil {
		logger.Fatal("error executing Router",
			zap.Error(err),
		)
	}
}
