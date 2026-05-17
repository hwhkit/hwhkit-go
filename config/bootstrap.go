package config

import "os"

type BootstrapConfig struct {
	ConfigDir string
	Env       string
}

func DefaultBootstrap() *BootstrapConfig {
	env := os.Getenv("HWHKIT_ENV")
	if env == "" {
		env = "dev"
	}
	dir := os.Getenv("HWHKIT_CONFIG_DIR")
	if dir == "" {
		dir = "config"
	}
	return &BootstrapConfig{ConfigDir: dir, Env: env}
}
