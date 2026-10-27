package config

import "os"

const defaultHTTPAddr = ":9021"

// Config is the process configuration loaded from the environment.
type Config struct {
	HTTPAddr string
}

// Load reads process configuration. HTTP_ADDR defaults to :9021.
// Database and Redis settings are intentionally unread until those clients exist.
func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = defaultHTTPAddr
	}
	return Config{HTTPAddr: addr}
}
