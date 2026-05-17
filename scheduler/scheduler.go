// Package scheduler wraps riverqueue/river for durable PG-backed cron and one-shot jobs.
package scheduler

import (
	"context"
	"errors"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/integration/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const Key = "scheduler"

type Scheduler struct {
	client *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
}

func (s *Scheduler) Client() *river.Client[pgx.Tx] { return s.client }

func New(pool *pgxpool.Pool, cfg config.SchedulerConfig, workers *river.Workers) (*Scheduler, error) {
	if pool == nil {
		return nil, errors.New("scheduler: pgxpool is nil")
	}
	queues := map[string]river.QueueConfig{
		cfgQueueName(cfg): {MaxWorkers: cfgMaxWorkers(cfg)},
	}
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  queues,
		Workers: workers,
	})
	if err != nil {
		return nil, coreerror.Integration(Key, coreerror.KindIntegration, err)
	}
	return &Scheduler{client: client, pool: pool}, nil
}

func (s *Scheduler) Start(ctx context.Context) error { return s.client.Start(ctx) }

func (s *Scheduler) Stop(ctx context.Context) error { return s.client.Stop(ctx) }

func cfgQueueName(c config.SchedulerConfig) string {
	if c.Queue == "" {
		return river.QueueDefault
	}
	return c.Queue
}

func cfgMaxWorkers(c config.SchedulerConfig) int {
	if c.MaxWorkers <= 0 {
		return 10
	}
	return c.MaxWorkers
}

// Provider implements core.IntegrationProvider so the scheduler lifecycle can be managed
// by the bootstrap pipeline. Register it AFTER the postgres provider.
type Provider struct {
	core.BaseProvider
	Workers *river.Workers
}

func NewProvider(workers *river.Workers) *Provider { return &Provider{Workers: workers} }

func (Provider) Key() string                        { return Key }
func (Provider) Enabled(c *config.AppConfig) bool  { return c.Scheduler.Enabled }

func (p *Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	h, ok := appctx.Get[postgres.Handle](app)
	if !ok {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured,
			"scheduler requires postgres provider to be registered and initialised first")
	}
	s, err := New(h.Pool(), c.Scheduler, p.Workers)
	if err != nil {
		return err
	}
	if err := s.Start(ctx); err != nil {
		return coreerror.Integration(Key, coreerror.KindIntegration, err)
	}
	appctx.Insert(app, s)
	app.InsertNamed(Key, s)
	return nil
}

func (Provider) Shutdown(ctx context.Context, app *appctx.Context) error {
	s, ok := appctx.Get[Scheduler](app)
	if !ok {
		return nil
	}
	return s.Stop(ctx)
}
