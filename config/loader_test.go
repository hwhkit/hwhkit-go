package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderEmbeddedDefaults(t *testing.T) {
	loader := DefaultLoader()
	res, err := loader.Load(context.Background(), &BootstrapConfig{ConfigDir: "/nonexistent", Env: "dev"})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Server.BindAddr != "0.0.0.0:8080" {
		t.Fatalf("expected default bind_addr, got %q", res.Config.Server.BindAddr)
	}
	if len(res.AppliedSources) == 0 || res.AppliedSources[0] != "embedded:default.toml" {
		t.Fatalf("expected embedded:default.toml first, got %v", res.AppliedSources)
	}
}

func TestLoaderEnvOverlay(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(envFile, []byte("[server]\nbind_addr = \"127.0.0.1:9090\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loader := DefaultLoader()
	res, err := loader.Load(context.Background(), &BootstrapConfig{ConfigDir: dir, Env: "test"})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Server.BindAddr != "127.0.0.1:9090" {
		t.Fatalf("env overlay not applied: %q", res.Config.Server.BindAddr)
	}
}

func TestLoaderEnvVarOverride(t *testing.T) {
	t.Setenv("HWHKIT__SERVER__BIND_ADDR", "0.0.0.0:7777")
	loader := DefaultLoader()
	res, err := loader.Load(context.Background(), &BootstrapConfig{ConfigDir: "/nonexistent", Env: "dev"})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Server.BindAddr != "0.0.0.0:7777" {
		t.Fatalf("ENV override not applied: %q", res.Config.Server.BindAddr)
	}
}
