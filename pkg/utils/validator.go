package utils

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const MaxFileSizeKey = "uploadMaxSize"

// Validate is the validator instance.
var Validate *validator.Validate

func init() {
	Validate = validator.New()
	Validate.RegisterValidationCtx("max-file-size", validateMaxFileSize, true)
}

// validateMaxFileSize validates the upload max file size.
func validateMaxFileSize(ctx context.Context, fl validator.FieldLevel) bool {
	uploadMaxSize := ctx.Value(MaxFileSizeKey).(int64)

	return fl.Field().Int() <= uploadMaxSize
}

// FormatValidatorErrors formats the validator.ValidationErrors to a string.
func FormatValidatorErrors(err error) []string {
	if err == nil {
		return nil
	}

	var errSlice []string
	for _, err := range err.(validator.ValidationErrors) {
		errMsg := fmt.Sprintf("%s field validation failed for tag '%s'", err.Field(), err.Tag())
		if err.Param() != "" {
			errMsg = fmt.Sprintf("%s: must satisfy condition '%s'", errMsg, err.Param())
		}
		errSlice = append(errSlice, fmt.Sprintf(errMsg))
	}

	return errSlice
}
