package config

import "testing"

func TestLoadDefaultsHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	cfg := Load()
	if cfg.HTTPAddr != ":9021" {
		t.Fatalf("HTTPAddr = %q, want :9021", cfg.HTTPAddr)
	}
}

func TestLoadUsesHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	cfg := Load()
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
}
