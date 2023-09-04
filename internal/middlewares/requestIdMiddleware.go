package middlewares

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestIdMiddleware adds an identifier to the x-request-id header.
func RequestIdMiddleware() gin.HandlerFunc {
	return requestid.New()
}
