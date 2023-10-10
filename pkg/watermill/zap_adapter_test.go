package watermill

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewZapLoggerAdapter(t *testing.T) {
	var config *zap.Logger

	assert.Implements(t, (*watermill.LoggerAdapter)(nil), NewZapLoggerAdapter(config))
}

func TestWith(t *testing.T) {
	fields := watermill.LogFields{
		"test": "test",
	}

	adapter := NewZapLoggerAdapter(logger.Logger)
	newAdapter := adapter.With(fields)
	assert.IsType(t, adapter, newAdapter)
}
