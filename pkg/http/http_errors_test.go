package http_utils

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRestError(t *testing.T) {
	msg := "message"
	description := []string{"description"}
	err := NewRestError(0, msg, description)

	assert.Implements(t, (*RestErr)(nil), err)
	assert.Equal(t, fmt.Sprintf("%d %s", err.Status(), msg), err.Error())
	assert.Equal(t, description, err.GetDescription())
	assert.Equal(t, 0, err.Status())
}

func TestNewBadRequestError(t *testing.T) {
	msg := "message"
	description := []string{"description"}
	err := NewBadRequestError(msg, description...)

	assert.Implements(t, (*RestErr)(nil), err)
	assert.Equal(t, fmt.Sprintf("%d %s", err.Status(), msg), err.Error())
	assert.Equal(t, description, err.GetDescription())
	assert.Equal(t, http.StatusBadRequest, err.Status())
}

func TestNewUnauthorizedError(t *testing.T) {
	msg := "message"
	description := []string{"description"}
	err := NewUnauthorizedError(msg, description...)

	assert.Implements(t, (*RestErr)(nil), err)
	assert.Equal(t, fmt.Sprintf("%d %s", err.Status(), msg), err.Error())
	assert.Equal(t, description, err.GetDescription())
	assert.Equal(t, http.StatusUnauthorized, err.Status())
}

func TestNewForbiddenError(t *testing.T) {
	msg := "message"
	description := []string{"description"}
	err := NewForbiddenError(msg, description...)

	assert.Implements(t, (*RestErr)(nil), err)
	assert.Equal(t, fmt.Sprintf("%d %s", err.Status(), msg), err.Error())
	assert.Equal(t, description, err.GetDescription())
	assert.Equal(t, http.StatusForbidden, err.Status())
}

func TestNewInternalServerError(t *testing.T) {
	msg := "message"
	description := []string{"description"}
	err := NewInternalServerError(msg, description...)

	assert.Implements(t, (*RestErr)(nil), err)
	assert.Equal(t, fmt.Sprintf("%d %s", err.Status(), msg), err.Error())
	assert.Equal(t, description, err.GetDescription())
	assert.Equal(t, http.StatusInternalServerError, err.Status())
}
