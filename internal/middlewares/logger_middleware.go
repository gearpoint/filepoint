package middlewares

import (
	"time"

	http_utils "github.com/gearpoint/filepoint/pkg/http"
	"github.com/gearpoint/filepoint/pkg/logger"
	ginzap "github.com/gin-contrib/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware sets the logger with custom config.
func LoggerMiddleware() gin.HandlerFunc {
	return ginzap.GinzapWithConfig(logger.Logger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/v1/health"},
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			fields = append(fields, zap.String("request_id", http_utils.GetRequestId(c)))

			return fields
		}),
	})
}
