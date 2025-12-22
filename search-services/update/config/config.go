package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type XKCD struct {
	URL         string        `yaml:"url" env:"XKCD_URL" env-default:"xkcd.com"`
	Concurrency int64         `yaml:"concurrency" env:"XKCD_CONCURRENCY" env-default:"1"`
	Timeout     time.Duration `yaml:"timeout" env:"XKCD_TIMEOUT" env-default:"10s"`
	Schedule    string        `yaml:"schedule" env:"XKCD_SCHEDULE" env-default:"0 9 * * 1,3,5"`
}

type Minio struct {
	RootUser       string `yaml:"root_user" env:"MINIO_ROOT_USER"`
	RootPassword   string `yaml:"root_password" env:"MINIO_ROOT_PASSWORD"`
	UseSSl         bool   `yaml:"use_ssl" env:"MINIO_USE_SSL" env-default:"false"`
	BucketName     string `yaml:"bucket_name" env:"MINIO_BUCKET_NAME"`
	ConnectAddress string `yaml:"connect_address" env:"MINIO_CONNECT_ADDRESS" env-default:"localhost:9000"`
	PublicAddress  string `yaml:"public_address" env:"MINIO_PUBLIC_ADDRESS" env-default:"localhost:9000"`
}

type Config struct {
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address       string `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"localhost:80"`
	XKCD          XKCD   `yaml:"xkcd"`
	DBAddress     string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	WordsAddress  string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:81"`
	BrokerAddress string `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"nats://localhost:4222"`
	Minio         Minio
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
