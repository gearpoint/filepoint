package http_utils

import (
	"mime/multipart"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// GetRequestId gets the request identifier from gin context.
func GetRequestId(ctx *gin.Context) string {
	return requestid.Get(ctx)
}

// GetConfigPath gets the config path for local or docker.
func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config/config-docker"
	}

	return "./config/config-local"
}

// GetIPAddress gets the user ip address.
func GetIPAddress(ctx *gin.Context) string {
	return ctx.Request.RemoteAddr
}

// ReadQueryParam gets the request body.
func ReadQueryParam(ctx *gin.Context, request interface{}) error {
	if err := ctx.Bind(request); err != nil {
		return err
	}
	return nil
}

// ReadRequest gets the request body.
func ReadRequest(ctx *gin.Context, request interface{}) error {
	if err := ctx.Bind(request); err != nil {
		return err
	}
	return nil
}

// ReadRequestFile reads a file from request and returns a FileHeader instance.
func ReadRequestFile(ctx *gin.Context, field string) (*multipart.FileHeader, error) {
	file, err := ctx.FormFile(field)
	if err != nil {
		return nil, err
	}

	return file, nil
}
