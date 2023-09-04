package logger

import (
	"context"
	"log"
	"os"

	"github.com/gearpoint/filepoint/pkg/utils"
	"go.uber.org/zap"
)

// loggerKeyType is used to identify the
type loggerKeyType int

const loggerKey loggerKeyType = iota

// The zap.Logger variable.
var Logger *zap.Logger

func init() {
	var err error

	switch utils.GetEnvironmentType() {
	case utils.Development:
		Logger, err = zap.NewDevelopment()
	case utils.Production:
		Logger, err = zap.NewProduction()
	}

	Logger.WithOptions(
		zap.Fields(zap.Int("pid", os.Getpid())),
	)

	if Logger == nil || err != nil {
		log.Fatal("Can't initialize zap logger: ", err)
	}

	defer Logger.Sync()
}

// NewContext creates a new context with the logger.
func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...))
}

// WithContext returns the logger with context inserted.
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(zap.Logger); ok {
		return &ctxLogger
	}

	return Logger
}

// Logger methods

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}
