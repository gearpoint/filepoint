package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/metrics"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthController is used for health checks.
type HealthController struct {
	awsRepository   *aws_repository.AWSRepository
	redisRepository *redis.RedisRepository
	logger          *zap.Logger
}

// NewHealthController creates a new HealthController
func NewHealthController(
	awsRepo *aws_repository.AWSRepository,
	redisRepo *redis.RedisRepository,
	logger *zap.Logger,
) *HealthController {
	return &HealthController{
		awsRepository:   awsRepo,
		redisRepository: redisRepo,
		logger:          logger,
	}
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status       string                 `json:"status"`
	Timestamp    string                 `json:"timestamp"`
	Dependencies map[string]DependencyHealth `json:"dependencies,omitempty"`
}

// DependencyHealth represents health of a single dependency
type DependencyHealth struct {
	Status  string `json:"status"` // up, down, degraded
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// HealthCheck godoc
// @Summary Basic health check
// @Description Returns a 200 OK response if service is running
// @Tags HealthCheck
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func (h *HealthController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthStatus{
		Status:    "up",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// LivenessCheck godoc
// @Summary Liveness probe
// @Description Returns 200 if application is alive (for Kubernetes liveness probe)
// @Tags HealthCheck
// @Produce plain
// @Success 200 {string} OK
// @Router /health/live [get]
func (h *HealthController) LivenessCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// ReadinessCheck godoc
// @Summary Readiness probe
// @Description Returns 200 if application is ready to serve traffic (checks all dependencies)
// @Tags HealthCheck
// @Produce json
// @Success 200 {object} HealthStatus
// @Failure 503 {object} HealthStatus
// @Router /health/ready [get]
func (h *HealthController) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dependencies := make(map[string]DependencyHealth)
	allHealthy := true

	// Check Redis
	redisHealth := h.checkRedis(ctx)
	dependencies["redis"] = redisHealth
	if redisHealth.Status != "up" {
		allHealthy = false
	}

	// Check S3
	s3Health := h.checkS3(ctx)
	dependencies["s3"] = s3Health
	if s3Health.Status != "up" {
		allHealthy = false
	}

	// Check DynamoDB
	dynamoHealth := h.checkDynamoDB(ctx)
	dependencies["dynamodb"] = dynamoHealth
	if dynamoHealth.Status != "up" {
		allHealthy = false
	}

	status := HealthStatus{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: dependencies,
	}

	if allHealthy {
		status.Status = "up"
		c.JSON(http.StatusOK, status)
	} else {
		status.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// checkRedis checks Redis connectivity
func (h *HealthController) checkRedis(ctx context.Context) DependencyHealth {
	start := time.Now()

	// Try a simple ping operation
	err := h.redisRepository.Client.Ping(ctx).Err()
	latency := time.Since(start)

	// Update metrics
	if err != nil {
		metrics.Get().DependencyUp.WithLabelValues("redis").Set(0)
		h.logger.Error("Redis health check failed", zap.Error(err))
		return DependencyHealth{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	metrics.Get().DependencyUp.WithLabelValues("redis").Set(1)
	return DependencyHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkS3 checks S3 connectivity
func (h *HealthController) checkS3(ctx context.Context) DependencyHealth {
	start := time.Now()

	// Try to list buckets (lightweight operation)
	_, err := h.awsRepository.S3Client.ListBuckets(ctx, nil)
	latency := time.Since(start)

	if err != nil {
		metrics.Get().DependencyUp.WithLabelValues("s3").Set(0)
		h.logger.Error("S3 health check failed", zap.Error(err))
		return DependencyHealth{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	metrics.Get().DependencyUp.WithLabelValues("s3").Set(1)
	return DependencyHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkDynamoDB checks DynamoDB connectivity
func (h *HealthController) checkDynamoDB(ctx context.Context) DependencyHealth {
	start := time.Now()

	// Try to list tables (lightweight operation)
	_, err := h.awsRepository.DynamoDBClient.ListTables(ctx, nil)
	latency := time.Since(start)

	if err != nil {
		metrics.Get().DependencyUp.WithLabelValues("dynamodb").Set(0)
		h.logger.Error("DynamoDB health check failed", zap.Error(err))
		return DependencyHealth{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	metrics.Get().DependencyUp.WithLabelValues("dynamodb").Set(1)
	return DependencyHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}
