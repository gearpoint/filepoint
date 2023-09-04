package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestHealthRoute(t *testing.T) {
	s := server.NewServer(&config.Config{
		Server: config.ServerConfig{},
		Redis:  config.RedisConfig{},
		S3:     config.S3{},
	}, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "v1/health", nil)
	s.Engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
