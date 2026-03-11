package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env"`

	GRPCConfig struct {
		Port    int           `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"grpc"`

	StorageConfig struct {
		Postgres struct {
			DSN            string
			MaxOpenConns   int32         `yaml:"max_open_conns"`
			MaxIdleConns   int32         `yaml:"max_idle_conns"`
			MaxIdleTime    time.Duration `yaml:"max_idle_time"`
			ConnAttempts   int32         `yaml:"conn_attempts"`
			BaseRetryDelay time.Duration `yaml:"base_retry_delay"`
			MaxRetryDelay  time.Duration `yaml:"max_retry_delay"`
		} `yaml:"postgres"`
	} `yaml:"storage"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("Load config path is failed")
	}

	//panic("Load config failed")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file %s does not exists", configPath))

	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("Parse config is failed: %s", err.Error()))
	}

	cfg.StorageConfig.Postgres.DSN = os.Getenv("DSN_POSTGRES")
	if cfg.StorageConfig.Postgres.DSN == "" {
		panic("Load DSN is failed")
	}

	return &cfg
}
