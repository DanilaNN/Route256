package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const pathToConfig = "config.yaml"

type Config struct {
	ConsumerGroup string   `yaml:"consumerGroup"`
	Brokers       []string `yaml:"kafkaBrokers"`
	Topic         string   `yaml:"kafkaTopic"`
	Env           string   `yaml:"environment"`
	MetricsHost   string   `yaml:"metricsHost"`
	GrpcPort      int      `yaml:"grpcPort"`
	HttpPort      int      `yaml:"httpPort"`
	PgConnStr     string   `yaml:"pgConnStr"`
	JaegerHost    string   `yaml:"jaegerHost"`
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
