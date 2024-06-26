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
	Server          ServerConfig
	Routes          Routes
	AWSConfig       AWSConfig
	StreamingConfig StreamingConfig
	RedisConfig     RedisConfig
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
	TableName   string
	Topic       string
	PoisonTopic string
	WebhookURL  string
	MaxRetries  int
}

// Routes defines the available routes.
type Routes map[Route]RouteConfig

// AWSConfig is the AWS configuration.
type AWSConfig struct {
	// Endpoint defines the AWS base URL where SDK will make API calls to. If empty will be resolved with the default URL.
	Endpoint           string
	Bucket             string
	Region             string
	CloudfrontCrtFile  string
	CloudfrontDist     string
	CloudfrontKeyId    string
	VideoLabelingTopic string
	RekognitionRole    string
}

// StreamingConfig contains the app streaming services configuration.
type StreamingConfig struct {
	MessagesPerSecond int64
	KafkaConfig       KafkaConfig
}

// KafkaConfig is the Kafka producer configuration.
type KafkaConfig struct {
	Brokers         []string
	MaxRetries      int
	MaxMessageBytes int
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
	v.AutomaticEnv()

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
