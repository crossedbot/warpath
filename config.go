package main

import (
	"github.com/crossedbot/common/golang/config"
)

type DatabaseConfig struct {
	Name               string `toml:"name"`
	Path               string `toml:"path"`
	MaxOpenConnections int    `toml:"max_open_connections"`
	MigrationsPath     string `toml:"migrations_path"`
	MigrationsEnv      string `toml:"migrations_environment"`
}

type LoggingConfig struct {
	Mode bool   `toml:"mode"`
	File string `toml:"file"`
}

type Config struct {
	SnapLength int            `toml:"snapshot_length"`
	Timeout    int            `toml:"timeout"`
	Filter     string         `toml:"filter"`
	Database   DatabaseConfig `toml:"database"`
	Logging    LoggingConfig  `toml:"logging"`
}

func Load(c *Config) error {
	return config.Load(c)
}
