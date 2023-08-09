package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const pathToConfig = "config.yaml"

type Config struct {
	Token       string `yaml:"token"`
	PgConnStr   string `yaml:"pgConnStr"`
	Port        int    `yaml:"port"`
	PortGrpc    int    `yaml:"portGrpc"`
	Env         string `yaml:"environment"`
	MetricsHost string `yaml:"metricsHost"`
	JaegerHost  string `yaml:"jaegerHost"`
	Services    struct {
		Loms            string `yaml:"loms"`
		ProductServ     string `yaml:"productService"`
		LomsGrpc        string `yaml:"lomsGrpc"`
		ProductServGrpc string `yaml:"productServiceGrpc"`
	} `yaml:"services"`
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
