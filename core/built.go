package core

import (
	"net/http"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/health"
	"github.com/hwhkit/hwhkit-go/core/shutdown"
)

type BuiltApplication struct {
	router                  http.Handler
	appCtx                  *appctx.Context
	cfg                     *config.AppConfig
	bootstrap               *config.BootstrapConfig
	appliedSources          []string
	initializedIntegrations []string
	degradedIntegrations    []string
	shutdownToken           *shutdown.Token
	healthRegistry          *health.Registry
	providers               []IntegrationProvider
	metricsRegistry         any
}

func (b *BuiltApplication) Router() http.Handler             { return b.router }
func (b *BuiltApplication) AppContext() *appctx.Context      { return b.appCtx }
func (b *BuiltApplication) Config() *config.AppConfig        { return b.cfg }
func (b *BuiltApplication) Bootstrap() *config.BootstrapConfig { return b.bootstrap }
func (b *BuiltApplication) AppliedSources() []string         { return b.appliedSources }
func (b *BuiltApplication) InitializedIntegrations() []string { return b.initializedIntegrations }
func (b *BuiltApplication) DegradedIntegrations() []string    { return b.degradedIntegrations }
func (b *BuiltApplication) Shutdown() *shutdown.Token        { return b.shutdownToken }
func (b *BuiltApplication) Health() *health.Registry         { return b.healthRegistry }
func (b *BuiltApplication) Providers() []IntegrationProvider { return b.providers }
func (b *BuiltApplication) MetricsRegistry() any             { return b.metricsRegistry }

// SetMetricsRegistry is called by the production server to attach the prometheus
// registry it created so that /metrics can serve from it. Hidden from godoc by
// not adding a comment on a separate exposed surface area; kept exported so
// the hwhkit facade (different module) can mutate this field.
func (b *BuiltApplication) SetMetricsRegistry(r any) { b.metricsRegistry = r }

func (b *BuiltApplication) SetRouter(r http.Handler) { b.router = r }
