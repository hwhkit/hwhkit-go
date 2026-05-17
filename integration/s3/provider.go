// Package s3 provides the hwhkit IntegrationProvider for S3-compatible storage via aws-sdk-go-v2.
package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
)

const Key = "s3"

type Handle struct {
	client *s3.Client
	bucket string
	region string
}

func (h *Handle) Client() *s3.Client { return h.client }
func (h *Handle) Bucket() string     { return h.bucket }
func (h *Handle) Region() string     { return h.region }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                          { return Key }
func (Provider) Enabled(c *config.AppConfig) bool    { return c.Integrations.Storage.S3.Enabled }
func (Provider) Required(c *config.AppConfig) bool   { return c.Integrations.Storage.S3.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	sc := &c.Integrations.Storage.S3
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(sc.Region),
	}
	if sc.AccessKey != "" && sc.SecretKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(sc.AccessKey, sc.SecretKey, ""),
		))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindInvalidConfig, err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if sc.Endpoint != "" {
			o.BaseEndpoint = aws.String(sc.Endpoint)
		}
		o.UsePathStyle = sc.UsePathStyle
	})

	if sc.Bucket != "" {
		probeCtx, cancel := context.WithTimeout(ctx, sc.Resilience.OpTimeout())
		defer cancel()
		if _, err := client.HeadBucket(probeCtx, &s3.HeadBucketInput{Bucket: aws.String(sc.Bucket)}); err != nil {
			return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
		}
	}

	h := &Handle{client: client, bucket: sc.Bucket, region: sc.Region}
	appctx.Insert(app, h)
	app.InsertNamed(Key, h)
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok || h.bucket == "" {
		return nil
	}
	probe := c.Integrations.Storage.S3.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if _, err := h.client.HeadBucket(pctx, &s3.HeadBucketInput{Bucket: aws.String(h.bucket)}); err != nil {
				return health.Result{Status: health.StatusUnhealthy, Detail: err.Error()}
			}
			return health.Result{Status: health.StatusHealthy}
		},
	}
}

func (Provider) Shutdown(context.Context, *appctx.Context) error {
	return nil
}

