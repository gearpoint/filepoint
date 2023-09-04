package middlewares

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	var value gin.HandlerFunc
	assert.IsType(t, value, LoggerMiddleware())
}
