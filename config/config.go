package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

// App config struct
type Config struct {
	Server ServerConfig
	Redis  RedisConfig
	S3     S3
}

// Server config struct
type ServerConfig struct {
	Environment       string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	SSL               bool
	CtxDefaultTimeout time.Duration
	Debug             bool
}

// Redis config
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

// AWS S3
type S3 struct {
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
