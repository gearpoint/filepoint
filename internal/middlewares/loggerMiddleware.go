// package middlewares have the API middlewares implementations.
package middlewares

import (
	"time"

	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gin-contrib/requestid"
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
		SkipPaths:  []string{"/health"},
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			fields = append(fields, zap.String("request_id", requestid.Get(c)))

			// if requestID := c.Writer.Header().Get("x-request-id"); requestID != "" {
			// 	fields = append(fields, zap.String("request_id", requestID))
			// }

			return fields
		}),
	})
}
