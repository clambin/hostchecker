package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
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

const defaultPort = 8080

func Read(r io.Reader) (*Config, error) {
	cfg := &Config{Port: defaultPort}
	if err := yaml.NewDecoder(r).Decode(cfg); err != nil {
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
