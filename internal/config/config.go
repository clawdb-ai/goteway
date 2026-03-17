package config

import (
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
}

// FromEnv loads config from env vars with safe defaults.
func FromEnv() (Config, error) {
	cfg := Config{
		Host:        env("GATEWAY_HOST", "127.0.0.1"),
		Port:        envInt("GATEWAY_PORT", 18789),
		ProtocolMin: envInt("GATEWAY_PROTOCOL_MIN", 1),
		ProtocolMax: envInt("GATEWAY_PROTOCOL_MAX", 3),
		AuthToken:   os.Getenv("GATEWAY_TOKEN"),
	}
	if cfg.ProtocolMin <= 0 {
		cfg.ProtocolMin = 1
	}
	if cfg.ProtocolMax < cfg.ProtocolMin {
		cfg.ProtocolMax = cfg.ProtocolMin
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

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
