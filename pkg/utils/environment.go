package utils

import "os"

const (
	// The EnvironmentKey defines the key that contains the environment config.
	EnvironmentKey string = "Environment"
)

// The EnvironmentType defines the app environment.
type EnvironmentType int64

// The app environment types.
const (
	Development EnvironmentType = iota
	Production
)

// GetEnv retrieves an environment variable.
func GetEnv(key string) string {
	return os.Getenv(key)
}

// GetEnvOrDefault retrieves an environment variable and uses a fallback value if empty.
func GetEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

// GetEnvironmentType returns the app environment.
func GetEnvironmentType() EnvironmentType {
	envType := GetEnv(EnvironmentKey)

	switch envType {
	case "production":
		return Production
	case "development":
		return Development
	default:
		return Production
	}
}
