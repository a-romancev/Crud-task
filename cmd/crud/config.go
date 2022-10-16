package main

import (
	"log"

	"github.com/a-romancev/crud_task/internal/event"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Mongo struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Database string `mapstructure:"database"`
}

type Config struct {
	ListenWebAddress string          `mapstructure:"listen"`
	LogLevel         string          `mapstructure:"loglevel"`
	Kafka            event.KafkaConf `mapstructure:"kafka"`
	Mongo            Mongo           `mapstructure:"mongo"`
	PublicKey        string          `mapstructure:"public_key"`
}

func (c Config) WithFile(confPath string) Config {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(confPath)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("config file not found in path")
		} else {
			log.Fatal("error while reading config file")
		}
	}
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatal("error while reading config file")
	}
	return c
}

func (c Config) Validate() error {
	if c.ListenWebAddress == "" {
		return errors.New("address not set")
	}
	if c.LogLevel == "" {
		return errors.New("loglevel not set")
	}
	if c.PublicKey == "" {
		return errors.New("publicKey not set")
	}
	if c.Mongo.Host == "" {
		return errors.New("mongoDB host not set")
	}
	if c.Mongo.Database == "" {
		return errors.New("mongoDB database not set")
	}
	if c.Mongo.User == "" {
		return errors.New("mongoDB user not set")
	}
	if c.Mongo.Password == "" {
		return errors.New("mongoDB password not set")
	}
	if len(c.Kafka.Servers) < 1 {
		return errors.New("kafka server not set")
	}
	if c.Kafka.Topic == "" {
		return errors.New("kafka topic not set")
	}
	return nil
}
