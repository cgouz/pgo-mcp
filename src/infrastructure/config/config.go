package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURI string
	MCPHost     string
	MCPPort     string
}

func Load() (Config, error) {

	cfg := Config{
		DatabaseURI: os.Getenv("DATABASE_URI"),
		MCPHost:     os.Getenv("MCP_HOST"),
		MCPPort:     os.Getenv("MCP_PORT"),
	}

	if cfg.DatabaseURI == "" {
		return Config{}, fmt.Errorf("DATABASE_URI is required")
	}

	if cfg.MCPHost == "" {
		cfg.MCPHost = "127.0.0.1"
	}

	if cfg.MCPPort == "" {
		cfg.MCPPort = "8001"
	}

	return cfg, nil
}
