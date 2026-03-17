package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config captures gateway runtime configuration.
type Config struct {
	Host        string
	Port        int
	ProtocolMin int
	ProtocolMax int
	AuthToken   string
	ShutdownSec int
}

// FromEnv loads config from env vars with safe defaults.
func FromEnv() (Config, error) {
	port, err := envInt("GATEWAY_PORT", 18789)
	if err != nil {
		return Config{}, err
	}
	minProtocol, err := envInt("GATEWAY_PROTOCOL_MIN", 1)
	if err != nil {
		return Config{}, err
	}
	maxProtocol, err := envInt("GATEWAY_PROTOCOL_MAX", 3)
	if err != nil {
		return Config{}, err
	}
	shutdownSec, err := envInt("GATEWAY_SHUTDOWN_SEC", 10)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Host:        env("GATEWAY_HOST", "127.0.0.1"),
		Port:        port,
		ProtocolMin: minProtocol,
		ProtocolMax: maxProtocol,
		AuthToken:   os.Getenv("GATEWAY_TOKEN"),
		ShutdownSec: shutdownSec,
	}
	if cfg.Host == "" {
		return Config{}, fmt.Errorf("GATEWAY_HOST must not be empty")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("GATEWAY_PORT must be between 1 and 65535")
	}
	if cfg.ProtocolMin <= 0 {
		return Config{}, fmt.Errorf("GATEWAY_PROTOCOL_MIN must be >= 1")
	}
	if cfg.ProtocolMax < cfg.ProtocolMin {
		return Config{}, fmt.Errorf("GATEWAY_PROTOCOL_MAX must be >= GATEWAY_PROTOCOL_MIN")
	}
	if cfg.ShutdownSec <= 0 {
		return Config{}, fmt.Errorf("GATEWAY_SHUTDOWN_SEC must be >= 1")
	}
	return cfg, nil
}

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func envInt(k string, def int) (int, error) {
	v := os.Getenv(k)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", k)
	}
	return n, nil
}
