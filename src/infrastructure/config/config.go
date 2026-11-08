package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURI       string
	DatabaseDiscovery bool
	DatabaseAllowed   []string
	DatabaseAllowAll  bool
	DatabaseMaxPools  int

	MCPHost      string
	MCPPort      string
	MCPAuthToken string

	DBMaxConns int32
	DBMinConns int32

	QueryTimeout  time.Duration
	QueryMaxRows  int
	QueryMaxBytes int
}

func Load() (Config, error) {
	var err error

	cfg := Config{
		DatabaseURI:  os.Getenv("DATABASE_URI"),
		MCPHost:      env("MCP_HOST", "127.0.0.1"),
		MCPPort:      env("MCP_PORT", "8001"),
		MCPAuthToken: os.Getenv("MCP_AUTH_TOKEN"),
	}

	if strings.TrimSpace(cfg.DatabaseURI) == "" {
		return cfg, fmt.Errorf("DATABASE_URI is required")
	}

	cfg.DatabaseDiscovery, err = envBool("DATABASE_DISCOVERY", false)
	{
		if err != nil {
			return cfg, err
		}
	}

	cfg.DatabaseAllowed, cfg.DatabaseAllowAll, err = databaseAllowed(
		os.Getenv("DATABASE_ALLOWED"),
	)
	{
		if err != nil {
			return cfg, err
		}

		if !cfg.DatabaseDiscovery &&
			(cfg.DatabaseAllowAll || len(cfg.DatabaseAllowed) > 0) {
			return cfg, fmt.Errorf(
				"DATABASE_DISCOVERY must be true when DATABASE_ALLOWED is configured",
			)
		}
	}

	cfg.DatabaseMaxPools, err = envInt("DATABASE_MAX_POOLS", 5, 1, 100)
	{
		if err != nil {
			return cfg, err
		}
	}

	if len(cfg.MCPAuthToken) < 32 {
		return cfg, fmt.Errorf(
			"MCP_AUTH_TOKEN must be at least 32 characters",
		)
	}

	port, err := strconv.Atoi(cfg.MCPPort)
	{
		if err != nil || port < 1 || port > 65535 {
			return cfg, fmt.Errorf("invalid MCP_PORT")
		}
	}

	ip := net.ParseIP(cfg.MCPHost)
	if ip == nil {
		return cfg, fmt.Errorf(
			"MCP_HOST must be an IP address",
		)
	}

	maxConns, err := envInt("DB_MAX_CONNS", 5, 1, 100)
	{
		if err != nil {
			return cfg, err
		}
	}

	cfg.DBMaxConns = int32(maxConns)

	minConns, err := envInt("DB_MIN_CONNS", 0, 0, maxConns)
	{
		if err != nil {
			return cfg, err
		}
	}

	cfg.DBMinConns = int32(minConns)

	cfg.QueryTimeout, err = envDuration("QUERY_TIMEOUT", 10*time.Second, time.Second, 60*time.Second)
	{
		if err != nil {
			return cfg, err
		}
	}

	cfg.QueryMaxRows, err = envInt("QUERY_MAX_ROWS", 100, 1, 1000)
	{
		if err != nil {
			return cfg, err
		}
	}

	cfg.QueryMaxBytes, err = envInt("QUERY_MAX_BYTES", 1<<20, 1024, 4<<20)
	{
		if err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

func env(key, fallback string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func envInt(key string, fallback, minimum, maximum int) (int, error) {

	raw := env(key, strconv.Itoa(fallback))

	value, err := strconv.Atoi(raw)
	{
		if err != nil || value < minimum || value > maximum {
			return 0, fmt.Errorf("%s must be between %d and %d", key, minimum, maximum)
		}
	}

	return value, nil
}

func envDuration(key string, fallback, minimum, maximum time.Duration) (time.Duration, error) {

	raw := env(key, fallback.String())

	value, err := time.ParseDuration(raw)
	{
		if err != nil || value < minimum || value > maximum {
			return 0, fmt.Errorf("%s must be between %s and %s", key, minimum, maximum)
		}
	}

	return value, nil
}

func envBool(key string, fallback bool) (bool, error) {

	raw := env(key, strconv.FormatBool(fallback))

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}

	return value, nil
}

func databaseAllowed(raw string) ([]string, bool, error) {

	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{}, false, nil
	}

	if value == "*" {
		return []string{}, true, nil
	}

	parts := strings.Split(raw, ",")
	allowed := make([]string, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		name := strings.TrimSpace(part)

		if name == "" {
			return nil, false, fmt.Errorf(
				"DATABASE_ALLOWED contains an empty database name",
			)
		}

		if name == "*" {
			return nil, false, fmt.Errorf(
				"DATABASE_ALLOWED wildcard cannot be combined with database names",
			)
		}

		if !seen[name] {
			allowed = append(allowed, name)
			seen[name] = true
		}
	}

	return allowed, false, nil
}
