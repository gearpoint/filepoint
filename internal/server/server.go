package server

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/middlewares"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gin-gonic/gin"
)

// ServerConfig contains the server configuration.
type ServerConfig struct {
	Config          *config.ServerConfig
	Routes          config.Routes
	Publisher       message.Publisher
	AWSRepository   *aws_repository.AWSRepository
	RedisRepository *redis.RedisRepository
}

// Server struct.
type Server struct {
	Engine          *gin.Engine
	config          *config.ServerConfig
	routes          config.Routes
	publisher       message.Publisher
	awsRepository   *aws_repository.AWSRepository
	redisRepository *redis.RedisRepository
}

// NewServer is the Server constructor.
func NewServer(serverConfig ServerConfig) *Server {
	return &Server{
		Engine:          gin.New(),
		config:          serverConfig.Config,
		routes:          serverConfig.Routes,
		publisher:       serverConfig.Publisher,
		awsRepository:   serverConfig.AWSRepository,
		redisRepository: serverConfig.RedisRepository,
	}
}

// getAddres returns a formatted address port.
func (s *Server) getAddres(port string) string {
	return fmt.Sprintf(":%v", port)
}

// Run starts the server.
func (s *Server) Run() error {
	var mode string
	if s.config.Debug {
		mode = gin.DebugMode
	} else {
		mode = gin.ReleaseMode
	}

	gin.SetMode(mode)

	s.Engine.Use(
		middlewares.RequestIdMiddleware(),
		middlewares.LoggerMiddleware(),
		gin.Recovery(),
	)

	s.MapHandlers()

	port := s.getAddres(s.config.Addr)
	s.Engine.Run(port)

	return nil
}
