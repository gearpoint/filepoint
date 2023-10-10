package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	var mode Mode

	mode = DevelopmentMode
	InitLogger(mode)

	assert.NotNil(t, Logger)
	assert.Equal(t, zap.DebugLevel, Logger.Level())

	mode = ProductionMode
	InitLogger(mode)

	assert.NotNil(t, Logger)
	assert.Equal(t, zap.InfoLevel, Logger.Level())
}

func TestWithContext(t *testing.T) {
	InitLogger(DevelopmentMode)

	logger := WithContext(context.Background())

	assert.NotNil(t, logger)
	assert.Equal(t, Logger, logger)

	InitLogger(DevelopmentMode)

	assert.NotEqual(t, Logger, logger)
}
