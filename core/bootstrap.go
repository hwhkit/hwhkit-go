package core

import (
	"context"
	"fmt"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	"github.com/hwhkit/hwhkit-go/core/runtime"
	"github.com/hwhkit/hwhkit-go/core/shutdown"
)

func Bootstrap(ctx context.Context, app Application, bs *config.BootstrapConfig) (*BuiltApplication, error) {
	return BootstrapWith(ctx, app, bs, config.DefaultLoader(), runtime.New(), nil)
}

func BootstrapWith(
	ctx context.Context,
	app Application,
	bs *config.BootstrapConfig,
	loader *config.Loader,
	features *runtime.Features,
	providers []IntegrationProvider,
) (*BuiltApplication, error) {
	if bs == nil {
		bs = config.DefaultBootstrap()
	}
	if loader == nil {
		loader = config.DefaultLoader()
	}
	if features == nil {
		features = runtime.New()
	}

	loaded, err := loader.Load(ctx, bs)
	if err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}
	cfg := loaded.Config

	if err := validateFeatureMapping(cfg, features); err != nil {
		return nil, err
	}

	shutdownToken := shutdown.New()
	healthRegistry := health.NewRegistry()

	appCtx := appctx.New()
	appctx.Insert(appCtx, shutdownToken)
	appCtx.InsertNamed("shutdown", shutdownToken)
	appctx.Insert(appCtx, healthRegistry)
	appCtx.InsertNamed("health", healthRegistry)

	var initialized, degraded []string
	var retained []IntegrationProvider

	for _, p := range providers {
		if !p.Enabled(cfg) {
			continue
		}
		key := p.Key()
		if !features.Contains(key) {
			return nil, coreerror.FeatureMismatch(key)
		}

		if err := p.Init(ctx, appCtx, cfg); err != nil {
			if p.Required(cfg) {
				return nil, err
			}
			degraded = append(degraded, key)
			continue
		}
		if check := p.HealthCheck(appCtx, cfg); check != nil {
			healthRegistry.Register(check)
		}
		initialized = append(initialized, key)
		retained = append(retained, p)
	}

	router, err := app.BuildRouter(ctx, appCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("build_router failed: %w", err)
	}

	return &BuiltApplication{
		router:                  router,
		appCtx:                  appCtx,
		cfg:                     cfg,
		bootstrap:               bs,
		appliedSources:          loaded.AppliedSources,
		initializedIntegrations: initialized,
		degradedIntegrations:    degraded,
		shutdownToken:           shutdownToken,
		healthRegistry:          healthRegistry,
		providers:               retained,
	}, nil
}

func validateFeatureMapping(cfg *config.AppConfig, features *runtime.Features) error {
	check := func(enabled bool, key string) error {
		if enabled && !features.Contains(key) {
			return coreerror.FeatureMismatch(key)
		}
		return nil
	}
	for _, c := range []struct {
		enabled bool
		key     string
	}{
		{cfg.Integrations.SQL.Postgres.Enabled, "postgres"},
		{cfg.Integrations.Redis.Enabled, "redis"},
		{cfg.Integrations.MongoDB.Enabled, "mongodb"},
		{cfg.Integrations.Messaging.NATS.Enabled, "nats"},
		{cfg.Integrations.Vector.Qdrant.Enabled, "qdrant"},
		{cfg.Integrations.Neo4j.Enabled, "neo4j"},
		{cfg.Integrations.Storage.S3.Enabled, "s3"},
		{cfg.Integrations.OSS.Enabled, "oss"},
	} {
		if err := check(c.enabled, c.key); err != nil {
			return err
		}
	}
	return nil
}
