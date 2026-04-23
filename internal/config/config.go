package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AuthToken string `yaml:"auth_token"`
	Port      string `yaml:"port"`
	UseTLS    bool   `yaml:"use_tls"`
}

func GetOrCreateConfig(defaultPort string) (*Config, error) {
	path := "config.yaml"
	
	// Priority 1: Environment Variables (Good for Docker)
	if envToken := os.Getenv("SHIFT_AUTH_TOKEN"); envToken != "" {
		return &Config{AuthToken: envToken, Port: defaultPort, UseTLS: false}, nil
	}

	// Priority 2: Config File
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	// Priority 3: Generate New (Initial Setup)
	token := make([]byte, 32)
	rand.Read(token)
	newKey := hex.EncodeToString(token)

	cfg := &Config{
		AuthToken: newKey,
		Port:      defaultPort,
		UseTLS:    false,
	}

	data, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(path, data, 0600)
	
	fmt.Println("✨ Generated new security config.yaml")
	return cfg, nil
}