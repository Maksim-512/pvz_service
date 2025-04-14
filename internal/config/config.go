package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP       HTTPConfig       `yaml:"http"`
	GRPC       GRPCConfig       `yaml:"grpc"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Database   DatabaseConfig   `yaml:"database"`
	Auth       AuthConfig       `yaml:"auth"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type HTTPConfig struct {
	Addr           string        `yaml:"address"`
	Port           string        `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

type GRPCConfig struct {
	Port string `yaml:"port"`
}

type PrometheusConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	ServerAddress      string `yaml:"server_address"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	Database           string `yaml:"name"`
	MaxOpenConnections int    `yaml:"max_open_conns"`
	MaxIdleConnections int    `yaml:"max_idle_conns"`
	ConnMaxLifetime    int    `yaml:"conn_max_lifetime"` // in minutes
}

type AuthConfig struct {
	SecretKey string `yaml:"secret_key"`
}

type LoggingConfig struct {
	Env string `yaml:"env"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
