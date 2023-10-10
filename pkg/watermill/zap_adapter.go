package watermill

import (
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

// ZapLoggerAdapter is a zap logger implementation.
type ZapLoggerAdapter struct {
	watermill.LoggerAdapter
	logger *zap.Logger
}

// NewZapLoggerAdapter creates a ZapLoggerAdapter.
func NewZapLoggerAdapter(logger *zap.Logger) ZapLoggerAdapter {
	return ZapLoggerAdapter{logger: logger}
}

// parseFields parses the watermill fields to zap fields.
func (z ZapLoggerAdapter) parseFields(fields watermill.LogFields) []zap.Field {
	var zapField []zap.Field

	for k, v := range fields {
		zapField = append(zapField, zap.Any(k, v))
	}

	return zapField
}

// Logger adapter methods

// Error parses the watermill fields and user zap.Error method.
func (z ZapLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	zapFields := z.parseFields(fields)
	zapFields = append(zapFields, zap.Error(err))
	z.logger.Error(msg, zapFields...)
}

// Info parses the watermill fields and user zap.Info method.
func (z ZapLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	zapFields := z.parseFields(fields)
	z.logger.Info(msg, zapFields...)
}

// Debug parses the watermill fields and user zap.Debug method.
func (z ZapLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	zapFields := z.parseFields(fields)
	z.logger.Debug(msg, zapFields...)
}

// Trace parses the watermill fields and user zap.Debug method.
func (z ZapLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	zapFields := z.parseFields(fields)
	z.logger.Debug(msg, zapFields...) // verify otel
}

// With parses the watermill fields and adds to the logger.
func (z ZapLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	zapFields := z.parseFields(fields)

	z.logger.WithOptions(
		zap.Fields(zapFields...),
	)

	return z
}
