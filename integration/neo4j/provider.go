// Package neo4j provides the hwhkit IntegrationProvider for Neo4j via the official Go driver v5.
package neo4j

import (
	"context"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const Key = "neo4j"

type Handle struct {
	driver neo.DriverWithContext
}

func (h *Handle) Driver() neo.DriverWithContext { return h.driver }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                         { return Key }
func (Provider) Enabled(c *config.AppConfig) bool   { return c.Integrations.Neo4j.Enabled }
func (Provider) Required(c *config.AppConfig) bool  { return c.Integrations.Neo4j.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	nc := &c.Integrations.Neo4j
	if nc.URI == "" {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured, "neo4j uri is empty")
	}
	driver, err := neo.NewDriverWithContext(nc.URI, neo.BasicAuth(nc.Username, nc.Password, ""))
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	pctx, cancel := context.WithTimeout(ctx, nc.Resilience.OpTimeout())
	defer cancel()
	if err := driver.VerifyConnectivity(pctx); err != nil {
		_ = driver.Close(ctx)
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	appctx.Insert(app, &Handle{driver: driver})
	app.InsertNamed(Key, &Handle{driver: driver})
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	probe := c.Integrations.Neo4j.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if err := h.driver.VerifyConnectivity(pctx); err != nil {
				return health.Result{Status: health.StatusUnhealthy, Detail: err.Error()}
			}
			return health.Result{Status: health.StatusHealthy}
		},
	}
}

func (Provider) Shutdown(ctx context.Context, app *appctx.Context) error {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	return h.driver.Close(ctx)
}
