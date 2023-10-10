package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("testGetEnv", "testWithValue")
	assert.Equal(t, "testWithValue", GetEnv("testGetEnv"))

	os.Unsetenv("testGetEnv")
	assert.Equal(t, "", GetEnv("testGetEnv"))
}

func TestGetEnvOrDefault(t *testing.T) {
	assert.Equal(t, "", GetEnvOrDefault("testGetEnv", ""))

	os.Setenv("testGetEnv", "testWithValue")
	assert.Equal(t, "testWithValue", GetEnvOrDefault("testGetEnv", "test"))

	os.Unsetenv("testGetEnv")
	assert.Equal(t, "test", GetEnvOrDefault("testGetEnv", "test"))
}
