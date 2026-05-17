package core

import (
	"context"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/health"
)

type IntegrationProvider interface {
	Key() string
	Enabled(cfg *config.AppConfig) bool
	Required(cfg *config.AppConfig) bool
	Init(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) error
	HealthCheck(app *appctx.Context, cfg *config.AppConfig) health.Check
	Shutdown(ctx context.Context, app *appctx.Context) error
}

// BaseProvider supplies default implementations so integration authors can embed it
// and override only the methods they need.
type BaseProvider struct{}

func (BaseProvider) Required(*config.AppConfig) bool { return true }

func (BaseProvider) HealthCheck(*appctx.Context, *config.AppConfig) health.Check { return nil }

func (BaseProvider) Shutdown(context.Context, *appctx.Context) error { return nil }
