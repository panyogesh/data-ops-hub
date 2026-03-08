package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.yaml.in/yaml/v2"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`

	// Align the auth structure as per config.yaml
	Auth struct {
		Domain string `yaml:"domain"`
		API    struct {
			Audience string `yaml:"audience"`
		} `yaml:"api"`
		M2M struct {
			ClientID     string `yaml:"client_id"`
			ClientSecret string `yaml:"client_secret"`
		} `yaml:"m2m"`
		Web struct {
			ClientID     string `yaml:"client_id"`
			ClientSecret string `yaml:"client_secret"`
			RedirectURI  string `yaml:"redirect_uri"`
		} `yaml:"web"`
	} `yaml:"auth"`
}

func LoadConfig(path string) (*Config, error) {
	// Load env variables from .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	// Replace the content of cfg files with matching env variables
	cfg.Auth.Domain = os.Getenv("AUTH0_DOMAIN")
	cfg.Auth.API.Audience = os.Getenv("AUTH0_API_IDENTIFIER_AUDIENCE")
	cfg.Auth.M2M.ClientID = os.Getenv("M2M_AUTH0_CLIENT_ID")
	cfg.Auth.M2M.ClientSecret = os.Getenv("M2M_AUTH0_CLIENT_SECRET")
	cfg.Auth.Web.ClientID = os.Getenv("WEB_AUTH0_CLIENT_ID")
	cfg.Auth.Web.ClientSecret = os.Getenv("WEB_AUTH0_CLIENT_SECRET")
	cfg.Auth.Web.RedirectURI = os.Getenv("WEB_AUTH0_REDIRECT_URI")

	return &cfg, nil
}
