package cfg

import (
	"fmt"
	"os"

	"github.com/Netflix/go-env"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig
	App    AppConfig
}

type ServerConfig struct {
	Addr       string `env:"ADDR,default=:8080"`
	ConfigPath string `env:"CONFIG_PATH,default=config.yaml"`
}

type AppConfig struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Path    string            `yaml:"path"`
	Target  string            `yaml:"target"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Load env
	if _, err := env.UnmarshalFromEnviron(&cfg.Server); err != nil {
		return nil, fmt.Errorf("load env config: %w", err)
	}

	// Load YAML
	data, err := os.ReadFile(cfg.Server.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read yaml config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg.App); err != nil {
		return nil, fmt.Errorf("parse yaml config: %w", err)
	}

	return cfg, nil
}

func (c *Config) FindRoute(path string) *Route {
	for i := range c.App.Routes {
		r := &c.App.Routes[i]
		if r.Path == path {
			return r
		}
	}
	return nil
}
