package http_utils

import (
	"fmt"
	"net/http"
)

// RestErr is the REST errors interface.
type RestErr interface {
	Error() string
	Status() int
	GetDescription() []string
}

// RestError contains the REST errors.
type RestError struct {
	status      int
	Message     string
	Description []string
}

// Error is the errors interface method.
func (e RestError) Error() string {
	return e.Message
}

// Status returns the error status.
func (e RestError) Status() int {
	return e.status
}

// Description returns the error Description.
func (e RestError) GetDescription() []string {
	return e.Description
}

// NewRestError returns a RestError instance.
func NewRestError(status int, err string, description []string) RestErr {
	return RestError{
		status:      status,
		Message:     fmt.Sprintf("%d %s", status, err),
		Description: description,
	}
}

// NewBadRequestError is the default 400 error.
func NewBadRequestError(message string, description ...string) RestErr {
	status := http.StatusBadRequest

	return NewRestError(
		status,
		message,
		description,
	)
}

// NewNotFoundError is the default 404 error.
func NewNotFoundError(message string, description ...string) RestErr {
	status := http.StatusNotFound

	return NewRestError(
		status,
		message,
		description,
	)
}

// NewUnauthorizedError is the default 401 error.
func NewUnauthorizedError(message string, description ...string) RestErr {
	status := http.StatusUnauthorized

	return NewRestError(
		status,
		message,
		description,
	)
}

// NewForbiddenError is the default 403 error.
func NewForbiddenError(message string, description ...string) RestErr {
	status := http.StatusForbidden

	return NewRestError(
		status,
		message,
		description,
	)
}

// NewInternalServerError is the default 500 error.
func NewInternalServerError(message string, description ...string) RestErr {
	status := http.StatusInternalServerError

	return NewRestError(
		status,
		message,
		description,
	)
}
