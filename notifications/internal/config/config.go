package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const pathToConfig = "config.yaml"

type Config struct {
	TelegramToken  string   `yaml:"telegramToken"`
	TelegramChatId uint64   `yaml:"telegramChatId"`
	ConsumerGroup  string   `yaml:"consumerGroup"`
	Brokers        []string `yaml:"kafkaBrokers"`
	Topic          string   `yaml:"kafkaTopic"`
	Env            string   `yaml:"environment"`
	MetricsHost    string   `yaml:"metricsHost"`
	JaegerHost     string   `yaml:"jaegerHost"`
	CacheCapacity  int      `yaml:"cacheCapacity"`
	HttpPort       string   `yaml:"httpPort"`
}

var AppConfig = Config{}

var once sync.Once

func Init() error {

	var initErr error = nil

	once.Do(func() {
		rawYaml, err := os.ReadFile(pathToConfig)
		if err != nil {
			initErr = fmt.Errorf("read config file: %w", err)
		}

		err = yaml.Unmarshal(rawYaml, &AppConfig)
		if err != nil {
			initErr = fmt.Errorf("parse config file: %w", err)
		}
	})

	return initErr
}
