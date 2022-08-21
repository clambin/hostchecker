package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
)

// Config is the configuration file for host checker
type Config struct {
	Debug   bool    `yaml:"debug,omitempty"`
	Port    int     `yaml:"port,omitempty"`
	Targets Targets `yaml:"targets"`
}

// Targets is the list of targets that need to be checked. Currently only HTTP(s) hosts are supported
type Targets struct {
	HTTP []HTTPTarget `yaml:"http"`
}

// An HTTPTarget contains details on the HTTP(s) host that needs to be checked.
type HTTPTarget struct {
	// Name of the host. Used for logging and Prometheus metrics
	Name string `yaml:"name"`
	// URL of the host
	URL string `yaml:"url"`
	// HTTP Method to use to contact the host
	Method string `yaml:"method"`
	// Codes is the list of expected HTTP status codes. If the host responds with an status code that is not in the list,
	// it will be considered to be down. If empty, defaults to HTTP 200.
	Codes ValidHTTPStatusCodes `yaml:"codes"`
}

// ValidHTTPStatusCodes is the list of expected HTTP status codes
type ValidHTTPStatusCodes []int

// IsValidCode returns true if the provided code is an expected status code
func (vc ValidHTTPStatusCodes) IsValidCode(code int) bool {
	for _, validCode := range vc {
		if code == validCode {
			return true
		}
	}
	return false
}

const defaultPort = 8080

// Read parses a Config structure from the provided Reader
func Read(r io.Reader) (*Config, error) {
	cfg := &Config{Port: defaultPort}
	if err := yaml.NewDecoder(r).Decode(cfg); err != nil {
		return nil, err
	}

	for index := range cfg.Targets.HTTP {
		if cfg.Targets.HTTP[index].Method == "" {
			cfg.Targets.HTTP[index].Method = http.MethodGet
		}
		if len(cfg.Targets.HTTP[index].Codes) == 0 {
			cfg.Targets.HTTP[index].Codes = ValidHTTPStatusCodes{http.StatusOK}
		}
	}
	return cfg, nil
}
