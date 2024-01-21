// Filepoint is the Gearpoint's file manager service. It's built for performance.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/api"
	config "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/server"
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

// @title Filepoint
// @description Filepoint is the Gearpoint's file manager service.
// @contact.name Luan Baggio
// @contact.url https://github.com/luabagg
// @contact.email luanbaggio0@gmail.com
// @BasePath /v1
func main() {
	godotenv.Load()

	envType := utils.GetEnvironmentType()
	initLogger(envType)

	logger.Info("starting Filepoint server...")

	flag.StringVar(&configFile, "config", "./config/config-local.yaml", "aaa")
	flag.Parse()

	cfg := getCfg(configFile)

	publisher, partitionKey := setUpPublisher(cfg)
	defer publisher.Close()

	awsRepository, err := aws_repository.NewAWSRepository(&cfg.AWSConfig, context.Background())
	if err != nil {
		logger.Fatal("cannot initialize storage client",
			zap.Error(err),
		)
	}
	logger.Info("AWS connected")

	redisRepository := redis.NewRedisRepository(&cfg.RedisConfig)
	defer redisRepository.Client.Close()
	logger.Info("Redis connected")

	var version string

	versionByte, err := os.ReadFile("VERSION")
	if err == nil {
		version = string(versionByte)
	}
	api.SwaggerInfo.Version = version

	logger.Info("Filepoint on!",
		zap.String("version", version),
	)

	s := server.NewServer(server.ServerConfig{
		Config:          &cfg.Server,
		Routes:          cfg.Routes,
		PartitionKey:    partitionKey,
		Publisher:       publisher,
		AWSRepository:   awsRepository,
		RedisRepository: redisRepository,
	})
	if err = s.Run(); err != nil {
		logger.Fatal("error starting server")
	}
}

func initLogger(envType utils.EnvironmentType) {
	switch envType {
	case utils.Development:
		logger.InitLogger(logger.DevelopmentMode)
	case utils.Production:
		logger.InitLogger(logger.ProductionMode)
	default:
		log.Fatal("error initializing logger")
	}
}

func getCfg(configFile string) *config.Config {
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

	return cfg
}

func setUpPublisher(cfg *config.Config) (message.Publisher, string) {
	var err error
	var publisher message.Publisher
	var partitionKey string

	switch utils.GetPubSubType() {
	case utils.Kafka:
		publisher, err = watermill.NewKafkaPublisher(&cfg.KafkaConfig)
		if err == nil {
			logger.Info("Kafka publisher connected successfully",
				zap.Any("brokers", cfg.KafkaConfig.Brokers),
			)
		}
	case utils.SQS:
		publisher, err = watermill.NewSQSPublisher(&cfg.SQSConfig)
		if err == nil {
			logger.Info("SQS publisher connected successfully",
				zap.Any("region", cfg.SQSConfig.AWSRegion),
			)
		}
	default:
		log.Fatal("error initializing the publisher")
	}

	if err != nil {
		logger.Fatal("error initializing the publisher",
			zap.Error(err),
		)
	}

	return publisher, partitionKey
}
