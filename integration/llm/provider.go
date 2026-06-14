package llm

import (
	"context"
	"time"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
)

// Key is the integration's stable identifier (logs, metrics, errors).
const Key = "llm"

// Provider implements the hwhkit IntegrationProvider contract.
type Provider struct{ core.BaseProvider }

// NewProvider returns a stateless provider.
func NewProvider() *Provider { return &Provider{} }

// Key returns the integration identifier.
func (Provider) Key() string { return Key }

// Enabled reads cfg.Integrations.LLM.Enabled.
func (Provider) Enabled(c *config.AppConfig) bool { return c.Integrations.LLM.Enabled }

// Required reads cfg.Integrations.LLM.Required.
func (Provider) Required(c *config.AppConfig) bool { return c.Integrations.LLM.Required }

// Init wires backends and stashes the *Handle under both the typed and
// named entries of the AppContext.
func (Provider) Init(_ context.Context, app *appctx.Context, c *config.AppConfig) error {
	lc := &c.Integrations.LLM
	op := lc.Resilience.OpTimeout()

	var backends []Backend
	if lc.Providers.Anthropic.IsConfigured() {
		backends = append(backends, NewAnthropic(lc.Providers.Anthropic.APIKey, lc.Providers.Anthropic.BaseURL, op))
	}
	for _, entry := range []struct {
		key   string
		creds config.LLMProviderCreds
	}{
		{"openai", lc.Providers.OpenAI},
		{"deepseek", lc.Providers.DeepSeek},
		{"moonshot", lc.Providers.Moonshot},
		{"ollama", lc.Providers.Ollama},
	} {
		if entry.creds.IsConfigured() {
			backends = append(backends, NewOpenAICompat(entry.key, entry.creds.APIKey, entry.creds.BaseURL, op))
		}
	}
	if len(backends) == 0 {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured,
			"no backends configured under [integrations.llm.providers.*]")
	}

	h := NewHandle(
		backends,
		lc.DefaultChatModel,
		lc.DefaultEmbeddingModel,
		nonZeroF32(lc.DefaultTemperature, 0.7),
		lc.DefaultMaxTokens,
		op,
	)
	appctx.Insert(app, h)
	app.InsertNamed(Key, h)
	return nil
}

// HealthCheck inherits the no-op from core.BaseProvider — LLM API
// pings cost money, so we deliberately don't probe at /health/ready.
// Adopters who want a smoke test should issue a low-token Chat() from
// a /diag endpoint instead.
//
// Shutdown also inherits from BaseProvider: the HTTP clients are owned
// by the Handle and released when the AppContext drops.

func nonZeroF32(v, fallback float32) float32 {
	if v > 0 {
		return v
	}
	return fallback
}

// _ ensures time.Duration is referenced (avoids unused-import in some
// build configurations).
var _ = time.Second
