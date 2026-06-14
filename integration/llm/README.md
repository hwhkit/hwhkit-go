# integration/llm

Multi-provider LLM client (chat + embeddings) for the hwhkit-go
bootstrap pipeline.

## What it gives you

- A uniform `*Handle` exposed via `appctx.Context`, no SDK lock-in.
- Backends dispatched by **model prefix**:
  - `anthropic/...` → Anthropic Messages API
  - `openai/...` → OpenAI Chat Completions
  - `deepseek/...` → OpenAI-compatible (DeepSeek host)
  - `moonshot/...` → OpenAI-compatible (Moonshot host)
  - `ollama/...` → OpenAI-compatible (Ollama host)
- Streaming via a `<-chan StreamChunk`.
- Embeddings on the OpenAI-compatible backends.
- Per-call timeout from `cfg.Integrations.LLM.Resilience.op_timeout_ms`.

## What it does NOT do

- **No SDK dependency.** Speaks `net/http` against the wire shape
  directly. Keeps the dependency surface near zero.
- **No env-var sourcing of secrets.** API keys come from
  `cfg.Integrations.LLM.Providers.*.api_key` so they go through your
  config layer + secrets manager.
- **No prompt-caching / function calling / vision** yet — v1 is
  text-only. The roadmap layers those on top of the same `*Handle`.

## Config

```toml
[integrations.llm]
enabled = true
default_chat_model      = "anthropic/claude-3-5-sonnet-20241022"
default_embedding_model = "openai/text-embedding-3-small"
default_temperature     = 0.7
default_max_tokens      = 4096

[integrations.llm.resilience]
op_timeout_ms = 30000

[integrations.llm.providers.anthropic]
api_key = "${ANTHROPIC_API_KEY}"

[integrations.llm.providers.openai]
api_key = "${OPENAI_API_KEY}"

[integrations.llm.providers.deepseek]
api_key = "${DEEPSEEK_API_KEY}"

[integrations.llm.providers.ollama]
# no api_key needed for local Ollama
base_url = "http://localhost:11434"
```

## Usage

```go
import (
    "context"

    "github.com/hwhkit/hwhkit-go/core/appctx"
    "github.com/hwhkit/hwhkit-go/integration/llm"
)

func handler(ctx context.Context, app *appctx.Context) (string, error) {
    h, _ := appctx.Get[llm.Handle](app)
    resp, err := h.Chat(ctx, []llm.ChatMessage{
        llm.System("You are a concise assistant."),
        llm.User("Hello?"),
    }, llm.ChatOptions{})
    if err != nil {
        return "", err
    }
    return resp.Content, nil
}
```

Streaming:

```go
ch, err := h.ChatStream(ctx, messages, llm.ChatOptions{})
if err != nil {
    return err
}
for chunk := range ch {
    switch {
    case chunk.Err != nil:
        return chunk.Err
    case chunk.Kind == llm.ChunkText:
        fmt.Print(chunk.Text)
    case chunk.Kind == llm.ChunkDone:
        fmt.Printf("\n[%s]\n", chunk.FinishReason)
    }
}
```

## Wiring it into bootstrap

```go
import "github.com/hwhkit/hwhkit-go/integration/llm"

// In your bootstrap setup, add the provider before run/serve:
providers := []core.IntegrationProvider{
    llm.NewProvider(),
    // ...other providers
}
```

## Tests

```sh
cd integration/llm
go test ./...
```

7 tests over standard library `httptest`:

- SplitModel edge cases
- Handle resolution errors (no prefix / unknown prefix / no default)
- OpenAI chat happy path, bad status, and SSE streaming
- OpenAI embeddings
- Anthropic chat with `content` blocks + `usage` parsing

No real API keys required.

## License

Dual-licensed under [MIT](../../LICENSE-MIT) or
[Apache-2.0](../../LICENSE-APACHE).
