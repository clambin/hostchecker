package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
)

type Config struct {
	Debug   bool    `yaml:"debug,omitempty"`
	Port    int     `yaml:"port,omitempty"`
	Targets Targets `yaml:"targets"`
}

type Targets struct {
	Http []HTTPTarget `yaml:"http"`
}

type HTTPTarget struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Method string `yaml:"method"`
	Codes  []int  `yaml:"codes"`
}

func GetFromFile(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config file read failed: %w", err)
	}
	return Get(content)
}

const defaultPort = 8080

func Get(content []byte) (*Config, error) {
	cfg := &Config{
		Port: defaultPort,
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}

	for index := range cfg.Targets.Http {
		if cfg.Targets.Http[index].Method == "" {
			cfg.Targets.Http[index].Method = http.MethodGet
		}
		if len(cfg.Targets.Http[index].Codes) == 0 {
			cfg.Targets.Http[index].Codes = []int{http.StatusOK}
		}
	}
	return cfg, nil
}
