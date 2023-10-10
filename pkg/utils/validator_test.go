package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type teststruct struct {
	Test string `validate:"required,gte=10"`
}

func TestFormatValidatorErrors(t *testing.T) {
	testData := teststruct{
		Test: "lessthan",
	}

	err := Validate.Struct(testData)
	assert.Error(t, err)

	fmtErr := FormatValidatorErrors(err)
	assert.IsType(t, []string{""}, fmtErr)
}
