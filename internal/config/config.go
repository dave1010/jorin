package config

import (
	"os"
)

// Config holds basic configuration values for the application. This is a
// minimal starter implementation used during the Phase 1 refactor. It is
// intentionally simple: callers can extend fields as needed and wiring is
// done from cmd/jorin.
type Config struct {
	Model string
	// More fields (paths, timeouts, policy defaults) can be added later.
}

// Load loads a configuration from environment variables and returns a
// Config with sensible defaults. It does not attempt to read XDG files in
// this initial implementation; that will be added in Phase 3.
func Load() *Config {
	c := &Config{Model: "gpt-5-mini"}
	if m := os.Getenv("JORIN_MODEL"); m != "" {
		c.Model = m
	}
	return c
}
