package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string               `yaml:"env"`
	DBConfig    MongoDBStorageConfig `yaml:"storage_config"`
	CacheConfig RedisCacheConfig     `yaml:"cache_config"`
	HttpServer  HttpServerConfig     `yaml:"http_server"`
}

type SQLiteStorageConfig struct {
	StoragePath string `yaml:"storage_path"`
}

type MongoDBStorageConfig struct {
	ConnectionString string        `yaml:"connection_string"`
	Timeout          time.Duration `yaml:"timeout"`
}

type MapCacheConfig struct {
	Capacity int `yaml:"capacity"`
}

type RedisCacheConfig struct {
	ConnectionString string        `yaml:"connection_string"`
	DB               int           `yaml:"db"`
	Timeout          time.Duration `yaml:"timeout"`
	Capacity         int           `yaml:"capacity"`
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
