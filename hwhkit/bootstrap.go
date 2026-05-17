// Package hwhkit is the facade module exposing RunAndServe and the standard provider chain.
package hwhkit

import (
	"context"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/runtime"
	"github.com/hwhkit/hwhkit-go/hwhkit/production"
	"github.com/hwhkit/hwhkit-go/observability"
)

type Option func(*options)

type options struct {
	loader    *config.Loader
	features  *runtime.Features
	providers []core.IntegrationProvider
}

func defaults() options {
	return options{
		loader:   config.DefaultLoader(),
		features: runtime.New(),
	}
}

func WithProvider(p core.IntegrationProvider) Option {
	return func(o *options) {
		o.providers = append(o.providers, p)
		o.features.Enable(p.Key())
	}
}

func WithLoader(l *config.Loader) Option {
	return func(o *options) { o.loader = l }
}

func WithFeature(key string) Option {
	return func(o *options) { o.features.Enable(key) }
}

func Run(ctx context.Context, app core.Application, bs *config.BootstrapConfig, opts ...Option) (*core.BuiltApplication, error) {
	o := defaults()
	for _, opt := range opts {
		opt(&o)
	}
	return core.BootstrapWith(ctx, app, bs, o.loader, o.features, o.providers)
}

func RunAndServe(ctx context.Context, app core.Application, bs *config.BootstrapConfig, opts ...Option) error {
	built, err := Run(ctx, app, bs, opts...)
	if err != nil {
		return err
	}
	logger := observability.InitLogging(built.Config().Observability.Log)
	if built.Config().Observability.OTel.Enabled {
		_, shutdown, err := observability.InitTracing(ctx, built.Config().Observability.OTel)
		if err != nil {
			logger.Warn("tracing init failed", "err", err)
		} else {
			defer func() { _ = shutdown(context.Background()) }()
		}
	}
	return production.Serve(ctx, built, logger)
}
