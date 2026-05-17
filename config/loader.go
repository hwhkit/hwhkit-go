package config

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

//go:embed default.toml
var defaultTOML []byte

const (
	envPrefix    = "HWHKIT__"
	envSeparator = "__"
)

type Loaded struct {
	Config         *AppConfig
	AppliedSources []string
}

type Loader struct {
	HTTPClient *http.Client
	validator  *validator.Validate
}

func DefaultLoader() *Loader {
	return &Loader{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		validator:  validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (l *Loader) Load(ctx context.Context, bs *BootstrapConfig) (*Loaded, error) {
	if bs == nil {
		bs = DefaultBootstrap()
	}
	k := koanf.New(".")
	var applied []string

	if err := k.Load(rawbytes.Provider(defaultTOML), toml.Parser()); err != nil {
		return nil, fmt.Errorf("default.toml: %w", err)
	}
	applied = append(applied, "embedded:default.toml")

	if bs.ConfigDir != "" && bs.Env != "" {
		path := filepath.Join(bs.ConfigDir, bs.Env+".toml")
		if _, err := os.Stat(path); err == nil {
			if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			applied = append(applied, "file:"+path)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}
	}

	envProvider := env.ProviderWithValue(envPrefix, ".", func(key, value string) (string, any) {
		trimmed := strings.TrimPrefix(key, envPrefix)
		dotted := strings.ToLower(strings.ReplaceAll(trimmed, envSeparator, "."))
		return dotted, value
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("env: %w", err)
	}
	if hasEnvOverride() {
		applied = append(applied, "env:"+envPrefix+"*")
	}

	remoteURL := k.String("remote.url")
	if remoteURL != "" {
		body, err := fetchRemote(ctx, l.HTTPClient, remoteURL, k.Int("remote.timeout_ms"))
		if err != nil {
			return nil, fmt.Errorf("remote %s: %w", remoteURL, err)
		}
		if err := k.Load(rawbytes.Provider(body), toml.Parser()); err != nil {
			return nil, fmt.Errorf("remote parse: %w", err)
		}
		applied = append(applied, "remote:"+remoteURL)
	}

	cfg := &AppConfig{}
	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if err := l.validator.StructCtx(ctx, cfg); err != nil {
		return nil, &ValidationError{Err: err}
	}

	return &Loaded{Config: cfg, AppliedSources: applied}, nil
}

func hasEnvOverride() bool {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, envPrefix) {
			return true
		}
	}
	return false
}

func fetchRemote(ctx context.Context, client *http.Client, url string, timeoutMs int) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	timeout := 5 * time.Second
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remote returned %d", resp.StatusCode)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	if e.Err == nil {
		return "validation failed"
	}
	return "config validation failed: " + e.Err.Error()
}

func (e *ValidationError) Unwrap() error { return e.Err }
