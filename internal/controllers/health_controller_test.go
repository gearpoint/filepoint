package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gearpoint/filepoint/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestHealthRoute(t *testing.T) {
	s := server.NewServer(server.ServerConfig{})

	s.MapHandlers()

	router := s.Engine

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/health", nil)
	assert.Nil(t, err)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
