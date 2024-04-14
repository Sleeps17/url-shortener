package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string               `yaml:"env"`
	DBConfig    mongoDBStorageConfig `yaml:"storage_config"`
	CacheConfig redisCacheConfig     `yaml:"cache_config"`
	HttpServer  httpServerConfig     `yaml:"http_server"`
}

type sqliteStorageConfig struct {
	StoragePath string        `yaml:"storage_path"`
	Timeout     time.Duration `yaml:"timeout"`
}

type mongoDBStorageConfig struct {
	ConnectionString string        `yaml:"connection_string"`
	DBName           string        `yaml:"db_name"`
	CollectionName   string        `yaml:"collection_name"`
	Timeout          time.Duration `yaml:"timeout"`
}

type mapCacheConfig struct {
	Capacity int `yaml:"capacity"`
}

type redisCacheConfig struct {
	ConnectionString string        `yaml:"connection_string"`
	DB               int           `yaml:"db"`
	Timeout          time.Duration `yaml:"timeout"`
	Capacity         int           `yaml:"capacity"`
}

type httpServerConfig struct {
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
