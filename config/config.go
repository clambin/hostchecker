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

func (t *Config) UnmarshalYAML(node *yaml.Node) error {
	type cfg2 Config
	c := cfg2{
		Port: defaultPort,
	}

	if err := node.Decode(&c); err != nil {
		return err
	}
	for index := range c.Targets.Http {
		if c.Targets.Http[index].Method == "" {
			c.Targets.Http[index].Method = http.MethodGet
		}
		if len(c.Targets.Http[index].Codes) == 0 {
			c.Targets.Http[index].Codes = []int{http.StatusOK}
		}
	}
	*t = Config(c)
	return nil
}

const defaultPort = 8080

func Read(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := yaml.NewDecoder(r).Decode(cfg)
	return cfg, err
}
