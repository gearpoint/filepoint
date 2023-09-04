package utils

import (
	"context"

	"github.com/go-playground/validator/v10"
)

// Validat is the validator instance.
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// ValidateStruct validates struct fields.
func ValidateStruct(ctx context.Context, s interface{}) error {
	return Validate.StructCtx(ctx, s)
}
