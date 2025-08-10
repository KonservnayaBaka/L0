package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Server   ServerConfig
	Kafka    KafkaConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Level string `yml:"level"`
}

type DatabaseConfig struct {
	Host     string `yml:"host"`
	Port     int    `yml:"port"`
	User     string `yml:"user"`
	Password string `yml:"password"`
	DBName   string `yml:"db_name"`
	SSLMode  string `yml:"ssl_mode"`
}

type ServerConfig struct {
	Port    int           `yml:"port"`
	Timeout time.Duration `yml:"timeout"`
}

type KafkaConfig struct {
	Brokers   []string `yml:"brokers"`
	Topic     string   `yml:"topic"`
	GroupID   string   `mapstructure:"group_id"`
	Partition int      `yml:"partition"`
}

type RedisConfig struct {
	Host     string        `yml:"host"`
	Port     int           `yml:"port"`
	Password string        `yml:"password"`
	DB       int           `yml:"db"`
	TTL      time.Duration `yml:"ttl"`
}

func MustLoad() *Config {
	configFileFlag := flag.String("config", "", "config file with path")
	flag.Parse()

	viper.AutomaticEnv()

	if *configFileFlag != "" {
		viper.SetConfigFile(*configFileFlag)
	} else {
		viper.SetConfigFile("./internal/config/config.yml")
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config file: %w", err))
	}

	return &cfg
}
