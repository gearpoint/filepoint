package utils

import (
	"os"
)

const (
	// The EnvironmentKey defines the key that contains the environment config.
	EnvironmentKey string = "ENVIRONMENT"

	// The AddrKey defines the key that contains the app address.
	AddrKey string = "FILEPOINT_ADDR"

	// The PubSubKey defines the publisher/subscriber to be used.
	PubSubKey string = "PUBSUB"

	// The CloudfrontKeyId defines the key that contains the Cloudfront key ID.
	CloudfrontKeyId string = "AWS_CLOUDFRONT_KEY_ID"
)

// The EnvironmentType defines the app environment.
type EnvironmentType int64

// The app environment types.
const (
	Development EnvironmentType = iota
	Production
)

// The PubSubType defines the app pub/sub.
type PubSubType int64

// The app environment types.
const (
	Kafka PubSubType = iota
	SQS
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

// GetPubSubType returns the app pub/sub.
func GetPubSubType() PubSubType {
	envType := GetEnv(PubSubKey)

	switch envType {
	case "kafka":
		return Kafka
	case "sqs":
		return SQS
	default:
		return SQS
	}
}
