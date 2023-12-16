package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		HTTP   `yaml:"http"`
		Redis  `yaml:"redis"`
		Logger `yaml:"logger"`
	}
	// HTTP -.
	HTTP struct {
		Host string `env-required:"true" yaml:"host" env:"HTTP_HOST"`
		Port uint16 `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}
	// Redis -.
	Redis struct {
		RedisUrl string `env-required:"true" yaml:"url" env:"REDIS_URL"`
	}
	// Logger -.
	Logger struct {
		Level string `env-required:"false" yaml:"level" env:"LOGGER_LEVEL"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
