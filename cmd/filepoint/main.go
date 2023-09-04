package main

import (
	"flag"

	config "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/server"
	"github.com/gearpoint/filepoint/pkg/aws"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

var (
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
	logger.Info("Starting Filepoint server...")

	godotenv.Load()

	flag.StringVar(&configFile, "config", "./config/config-local.yaml", "aaa")
	flag.Parse()

	viperConfig, err := config.LoadConfig(configFile)
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

	s := server.NewServer(cfg, redisClient, s3Client)
	if err = s.Run(); err != nil {
		logger.Fatal("Error starting server")
	}

	logger.Info("Filepoint on!")
}
