package server

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/middlewares"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
	//_ "github.com/gearpoint/filepoint/api"
)

const (
	maxHeaderBytes = 1 << 20
	ctxTimeout     = 5
)

// Server struct
type Server struct {
	Engine      *gin.Engine
	config      *config.ServerConfig
	redisClient *redis.Client
	s3Client    *s3.Client
}

// NewServer new server constructor
func NewServer(cfg *config.Config, redisClient *redis.Client, s3Client *s3.Client) *Server {
	return &Server{
		Engine:      gin.New(),
		config:      &cfg.Server,
		redisClient: redisClient,
		s3Client:    s3Client,
	}
}

// getAddres returns a formatted address port.
func (s *Server) getAddres(port string) string {
	return fmt.Sprintf(":%v", port)
}

// Run starts the server.
func (s *Server) Run() error {
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
