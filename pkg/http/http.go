package http

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/AleksK1NG/api-mc/pkg/sanitize"
)

// GetRequestID gets the request identifier from gin context.
func GetRequestID(c *gin.Context) string {
	return requestid.Get(c)
}

// ReqIDCtxKey is a key used for the Request ID in context.
type ReqIDCtxKey struct{}

// GetCtxWithReqID gets the context with timeout and request id from gin context.
func GetCtxWithReqID(c *gin.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*15)
	ctx = context.WithValue(ctx, ReqIDCtxKey{}, GetRequestID(c))
	return ctx, cancel
}

// GetRequestCtx gets the context with request id.
func GetRequestCtx(c *gin.Context) context.Context {
	return context.WithValue(c.Request.Context(), ReqIDCtxKey{}, GetRequestID(c))
}

// GetConfigPath gets the config path for local or docker.
func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config/config-docker"
	}

	return "./config/config-local"
}

// GetIPAddress gets the user ip address.
func GetIPAddress(c *gin.Context) string {
	return c.Request.RemoteAddr
}

// Read request body and validate
func ReadRequest(ctx echo.Context, request interface{}) error {
	if err := ctx.Bind(request); err != nil {
		return err
	}
	return utils.Validate.StructCtx(ctx.Request().Context(), request)
}

func ReadImage(ctx echo.Context, field string) (*multipart.FileHeader, error) {
	image, err := ctx.FormFile(field)
	if err != nil {
		return nil, errors.WithMessage(err, "ctx.FormFile")
	}

	// Check content type of image
	if err = utils.CheckImageContentType(image); err != nil {
		return nil, err
	}

	return image, nil
}

// Read sanitize and validate request
func SanitizeRequest(ctx *gin.Context, request interface{}) error {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	defer ctx.Request.Body.Close()

	sanBody, err := sanitize.SanitizeJSON(body)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)

		return err
	}

	if err = json.Unmarshal(sanBody, request); err != nil {
		return err
	}

	return utils.Validate.StructCtx(ctx.Request.Context(), request)
}
