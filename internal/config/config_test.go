package config

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("GATEWAY_HOST", "")
	t.Setenv("GATEWAY_PORT", "")
	t.Setenv("GATEWAY_PROTOCOL_MIN", "")
	t.Setenv("GATEWAY_PROTOCOL_MAX", "")
	t.Setenv("GATEWAY_TOKEN", "")
	t.Setenv("GATEWAY_SHUTDOWN_SEC", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("expected defaults to load, got err=%v", err)
	}
	if cfg.Host == "" {
		t.Fatal("expected default host")
	}
	if cfg.Port <= 0 {
		t.Fatal("expected positive default port")
	}
	if cfg.ShutdownSec <= 0 {
		t.Fatal("expected positive default shutdown sec")
	}
}

func TestFromEnvInvalidPort(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "abc")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_PORT")
	}
}

func TestFromEnvInvalidProtocolRange(t *testing.T) {
	t.Setenv("GATEWAY_PROTOCOL_MIN", "3")
	t.Setenv("GATEWAY_PROTOCOL_MAX", "2")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid protocol range")
	}
}
