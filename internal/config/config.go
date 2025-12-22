package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Upstream UpstreamConfig `mapstructure:"upstream"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Limits   LimitsConfig   `mapstructure:"limits"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LimitsConfig struct {
	MaxBodySize      int `mapstructure:"max_body_size"`
	MaxBatchItems    int `mapstructure:"max_batch_items"`
	MaxBatchResponse int `mapstructure:"max_batch_response"`
}

type UpstreamConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8545)
	v.SetDefault("upstream.url", "http://localhost:8546")
	v.SetDefault("upstream.timeout", "30s")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("limits.max_body_size", 5242880)
	v.SetDefault("limits.max_batch_items", 100)
	v.SetDefault("limits.max_batch_response", 25000000)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
