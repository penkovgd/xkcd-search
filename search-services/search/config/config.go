package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel      string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address       string        `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"localhost:8080"`
	WordsAddress  string        `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:8081"`
	DBAddress     string        `yaml:"db_address" env:"DB_ADDRESS"`
	IndexTTL      time.Duration `yaml:"index_ttl" env:"INDEX_TTL" env-default:"20s"`
	BrokerAddress string        `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"nats://localhost:4222"`
}

func MustLoad(path string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return cfg
}
