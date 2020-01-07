package config

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	builder     *Builder
	once        sync.Once
	defaultPath = "config.toml"
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

type Builder struct {
	Path string
}

var build = func() *Builder {
	once.Do(func() {
		builder = &Builder{
			Path: defaultPath,
		}
	})
	return builder
}()

func Path(path string) {
	build.Path = filepath.Clean(path)
}

func Load(config *Config) error {
	b, err := ioutil.ReadFile(build.Path)
	if err != nil {
		return err
	}
	_, err = toml.Decode(string(b), config)
	return err
}
