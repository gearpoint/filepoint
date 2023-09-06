package main

import (
	"flag"

	config "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/server"
	"github.com/gearpoint/filepoint/pkg/aws"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	// configFile is the main configuration filepath.
	configFile string
)

// @title Filepoint
// @version 1.0
// @description Filepoint is the Gearpoint's file manager service.
// @contact.name Luan Baggio
// @contact.url https://github.com/luabagg
// @contact.email luanbaggio0@gmail.com
// @BasePath /v1
func main() {
	godotenv.Load()

	envType := utils.GetEnvironmentType()

	logger.InitLogger(envType)
	logger.Info("Starting Filepoint server...")

	flag.StringVar(&configFile, "config", "./config/config-docker.yaml", "aaa")
	flag.Parse()

	viperConfig, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Fatal("Error initializing config",
			zap.Error(err),
		)
	}

	cfg, err := config.ParseConfig(viperConfig)
	if err != nil {
		logger.Fatal("Error getting config",
			zap.Error(err),
		)
	}

	redisClient := redis.NewRedisClient(&cfg.Redis)
	defer redisClient.Close()
	logger.Info("Redis connected")

	s3Client, err := aws.NewS3Client(&cfg.S3)
	if err != nil {
		logger.Fatal("AWS Client init error",
			zap.Error(err),
		)
	}
	logger.Info("AWS S3 connected")

	logger.Info("Filepoint on!")

	s := server.NewServer(cfg, redisClient, s3Client)
	if err = s.Run(); err != nil {
		logger.Fatal("Error starting server")
	}
}
