// Package coreerror defines the typed error model used across hwhkit-go.
package coreerror

import (
	"errors"
	"fmt"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindTimeout
	KindConnectFailed
	KindInvalidURL
	KindAuthFailed
	KindNotConfigured
	KindFeatureMismatch
	KindInvalidConfig
	KindIntegration
)

func (k Kind) String() string {
	switch k {
	case KindTimeout:
		return "timeout"
	case KindConnectFailed:
		return "connect_failed"
	case KindInvalidURL:
		return "invalid_url"
	case KindAuthFailed:
		return "auth_failed"
	case KindNotConfigured:
		return "not_configured"
	case KindFeatureMismatch:
		return "feature_mismatch"
	case KindInvalidConfig:
		return "invalid_config"
	case KindIntegration:
		return "integration"
	default:
		return "unknown"
	}
}

type IntegrationError struct {
	Key  string
	Kind Kind
	Err  error
}

func (e *IntegrationError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("integration %q: %s", e.Key, e.Kind)
	}
	return fmt.Sprintf("integration %q: %s: %s", e.Key, e.Kind, e.Err)
}

func (e *IntegrationError) Unwrap() error { return e.Err }

func Integration(key string, kind Kind, err error) *IntegrationError {
	return &IntegrationError{Key: key, Kind: kind, Err: err}
}

func IntegrationMsg(key string, kind Kind, msg string) *IntegrationError {
	return &IntegrationError{Key: key, Kind: kind, Err: errors.New(msg)}
}

func FeatureMismatch(feature string) *IntegrationError {
	return &IntegrationError{
		Key:  feature,
		Kind: KindFeatureMismatch,
		Err:  fmt.Errorf("config enables integration %q but no provider registered", feature),
	}
}

func InvalidConfig(msg string) error {
	return &IntegrationError{Key: "config", Kind: KindInvalidConfig, Err: errors.New(msg)}
}
