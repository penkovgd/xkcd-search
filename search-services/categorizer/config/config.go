package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Ollama struct {
	URL   string `yaml:"url" env:"OLLAMA_URL" env-default:"http://localhost:11434"`
	Model string `yaml:"model" env:"OLLAMA_MODEL"`
}

type Config struct {
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address       string `yaml:"categorizer_address" env:"CATEGORIZER_ADDRESS" env-default:"localhost:8080"`
	BrokerAddress string `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"nats://localhost:4222"`
	Concurrency   int    `yaml:"concurrency" env:"CONRURRENCY" env-default:"10"`
	Ollama        Ollama
}

func MustLoad(path string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return cfg
}
