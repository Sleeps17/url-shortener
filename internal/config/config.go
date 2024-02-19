package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string           `yaml:"env"`
	StoragePath string           `yaml:"storage_path"`
	HttpServer  HttpServerConfig `yaml:"http_server"`
}

type HttpServerConfig struct {
	Port        string        `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")

	if path == "" {
		panic("CONFIG_PATH is not set")
	}

	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
