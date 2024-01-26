package config

import (
	"errors"
	"time"

	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/spf13/viper"
)

type Route string

const (
	Upload Route = "upload"
)

// Config is the app main config struct.
type Config struct {
	Server      ServerConfig
	Routes      Routes
	AWSConfig   AWSConfig
	KafkaConfig KafkaConfig
	SQSConfig   SQSConfig
	RedisConfig RedisConfig
}

// ServerConfig is the server configuration struct.
type ServerConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	SSL               bool
	CtxDefaultTimeout time.Duration
	Debug             bool
}

// Route config is the routes configuration.
type RouteConfig struct {
	Topic      string
	WebhookURL string
}

// Routes defines the available routes.
type Routes map[Route]RouteConfig

// AWSConfig config is the AWS configuration.
type AWSConfig struct {
	Bucket             string
	Region             string
	CloudfrontDist     string
	CloudfrontKeyId    string
	VideoLabelingTopic string
	RekognitionRole    string
}

// todo: fix messaging config

// KafkaConfig is the Kafka producer configuration.
type KafkaConfig struct {
	Brokers           []string
	MessagesPerSecond int64
	MaxMessageBytes   int
	MaxRetries        int
}

// SQSConfig is the SQS producer configuration.
type SQSConfig struct {
	AWSRegion string
}

// RedisConfig config is the Redis configuration.
type RedisConfig struct {
	Addr         string
	MinIdleConns int
	PoolSize     int
	PoolTimeout  int
	Username     string
	Password     string
}

// LoadConfig loads file from given path.
func LoadConfig(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AddConfigPath(".")

	v.SetDefault("Server.Addr", utils.GetEnv(utils.AddrKey))
	v.SetDefault("AWSConfig.CloudfrontKeyId", utils.GetEnv(utils.CloudfrontKeyId))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("Config file not found")
		}
		return nil, err
	}

	return v, nil
}

// Parse returns the parsed yaml content from the given file.
// The interface must match the file contents.
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
