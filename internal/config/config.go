package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string               `yaml:"env"`
	DBConfig    MongoDBStorageConfig `yaml:"storage_config"`
	CacheConfig RedisCacheConfig     `yaml:"cache_config"`
	HttpServer  HttpServerConfig     `yaml:"http_server"`
}

type SqliteStorageConfig struct {
	StoragePath string        `yaml:"storage_path"`
	Timeout     time.Duration `yaml:"timeout"`
}

type MongoDBStorageConfig struct {
	ConnectionString string        `yaml:"connection_string"`
	DBName           string        `yaml:"db_name"`
	CollectionName   string        `yaml:"collection_name"`
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
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		panic("CONFIG_PATH is not set")
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}

func MustLoadByPath(configPath string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
