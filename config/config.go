package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

// Config is the app main config struct.
type Config struct {
	Server ServerConfig
	Redis  RedisConfig
	AWS    AWS
}

// ServerConfig is the server configuration struct.
type ServerConfig struct {
	Environment       string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	SSL               bool
	CtxDefaultTimeout time.Duration
	Debug             bool
}

// RedisConfig is the Redis configuration.
type RedisConfig struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        string
	RedisDefaultdb string
	MinIdleConns   int
	PoolSize       int
	PoolTimeout    int
	Password       string
	DB             int
}

// AWS is the Amazon Web Services configuration.
type AWS struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// LoadConfig loads file from given path.
func LoadConfig(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AddConfigPath(".")

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
