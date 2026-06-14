package llm

import (
	"context"
	"strings"
)

// Backend is the per-provider client. One Backend instance is bound to
// one host + one set of credentials. The Handle multiplexes across N
// backends by model prefix.
type Backend interface {
	// Key returns the routing prefix this backend serves (e.g. "openai").
	Key() string
	// Chat runs a non-streaming completion. model is the prefix-stripped
	// name.
	Chat(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (*ChatResponse, error)
	// ChatStream returns a channel that yields chunks until either a
	// Done chunk arrives, the context is cancelled, or the channel
	// closes. The channel always closes; the consumer should not close it.
	ChatStream(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (<-chan StreamChunk, error)
	// Embed turns text inputs into vectors. Backends that don't support
	// embeddings return KindInvalidRequest.
	Embed(ctx context.Context, model string, texts []string, opts EmbedOptions) (*EmbedResponse, error)
}

// SplitModel splits "prefix/name" into ("prefix", "name"). Returns
// ("", "", false) if there's no slash.
func SplitModel(full string) (prefix, name string, ok bool) {
	idx := strings.IndexByte(full, '/')
	if idx <= 0 || idx == len(full)-1 {
		return "", "", false
	}
	return full[:idx], full[idx+1:], true
}
