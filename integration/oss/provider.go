// Package oss provides the hwhkit IntegrationProvider for Aliyun OSS.
package oss

import (
	"context"
	"time"

	ali "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
)

const Key = "oss"

type Handle struct {
	client *ali.Client
	bucket string
}

func (h *Handle) Client() *ali.Client { return h.client }
func (h *Handle) Bucket() string      { return h.bucket }

func (h *Handle) GetBucket() (*ali.Bucket, error) { return h.client.Bucket(h.bucket) }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                         { return Key }
func (Provider) Enabled(c *config.AppConfig) bool   { return c.Integrations.OSS.Enabled }
func (Provider) Required(c *config.AppConfig) bool  { return c.Integrations.OSS.Required }

func (Provider) Init(_ context.Context, app *appctx.Context, c *config.AppConfig) error {
	oc := &c.Integrations.OSS
	if oc.Endpoint == "" || oc.AccessKey == "" || oc.SecretKey == "" {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured, "oss endpoint/access_key/secret_key required")
	}
	client, err := ali.New(oc.Endpoint, oc.AccessKey, oc.SecretKey,
		ali.Timeout(int64(oc.Resilience.ConnectTimeout().Seconds()), int64(oc.Resilience.OpTimeout().Seconds())),
	)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}
	if oc.Bucket != "" {
		if _, err := client.GetBucketInfo(oc.Bucket); err != nil {
			return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
		}
	}
	appctx.Insert(app, &Handle{client: client, bucket: oc.Bucket})
	app.InsertNamed(Key, &Handle{client: client, bucket: oc.Bucket})
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, _ *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok || h.bucket == "" {
		return nil
	}
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			done := make(chan error, 1)
			go func() {
				_, err := h.client.GetBucketInfo(h.bucket)
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					return health.Result{Status: health.StatusUnhealthy, Detail: err.Error()}
				}
				return health.Result{Status: health.StatusHealthy}
			case <-ctx.Done():
				return health.Result{Status: health.StatusUnhealthy, Detail: "probe timeout"}
			case <-time.After(2 * time.Second):
				return health.Result{Status: health.StatusUnhealthy, Detail: "probe timeout"}
			}
		},
	}
}
