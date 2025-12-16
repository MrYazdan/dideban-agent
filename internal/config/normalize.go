package config

import "strings"

// normalizeConfig normalizes configuration values into
// a canonical form for internal use.
func normalizeConfig(cfg *Config) {
	cfg.Log.Level = strings.ToLower(cfg.Log.Level)
}
