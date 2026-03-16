package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env"`

	Storage struct {
		MongoDB struct {
			DSN    string `yaml:"dsn"`
			DBName string `yaml:"db_name"`
		} `yaml:"mongodb"`
	} `yaml:"storage"`

	RabbitMQ struct {
		URL                   string        `yaml:"url"`
		Exchange              string        `yaml:"exchange"`
		LargeTranslationQueue string        `yaml:"large_translation_queue"`
		PrefetchCount         int           `yaml:"prefetch_count"`
		ReconnectDelay        time.Duration `yaml:"reconnect_delay"`
	} `yaml:"rabbit_mq"`
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

	cfg.Storage.MongoDB.DSN = os.Getenv("DSN_MONGO")
	if cfg.Storage.MongoDB.DSN == "" {
		panic("Load DSN is failed")
	}

	return &cfg
}
