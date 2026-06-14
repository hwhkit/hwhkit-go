package llm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

// Handle is the user-facing client. Safe for concurrent use; cheap to
// pass by value of the *Handle pointer.
type Handle struct {
	backends              map[string]Backend
	defaultChatModel      string
	defaultEmbeddingModel string
	defaultTemperature    float32
	defaultMaxTokens      int
	opTimeout             time.Duration
}

// NewHandle wires up a Handle from a set of backends. Typically called
// from Provider.Init; users don't construct this directly.
func NewHandle(backends []Backend, defaultChatModel, defaultEmbeddingModel string, defaultTemperature float32, defaultMaxTokens int, opTimeout time.Duration) *Handle {
	m := make(map[string]Backend, len(backends))
	for _, b := range backends {
		m[b.Key()] = b
	}
	return &Handle{
		backends:              m,
		defaultChatModel:      defaultChatModel,
		defaultEmbeddingModel: defaultEmbeddingModel,
		defaultTemperature:    defaultTemperature,
		defaultMaxTokens:      defaultMaxTokens,
		opTimeout:             opTimeout,
	}
}

// OpTimeout returns the configured per-call timeout.
func (h *Handle) OpTimeout() time.Duration { return h.opTimeout }

// DefaultChatModel returns the prefix-qualified default chat model.
func (h *Handle) DefaultChatModel() string { return h.defaultChatModel }

// DefaultEmbeddingModel returns the prefix-qualified default embedding
// model.
func (h *Handle) DefaultEmbeddingModel() string { return h.defaultEmbeddingModel }

// BackendKeys returns the sorted set of wired backend prefixes — for
// observability and startup logs.
func (h *Handle) BackendKeys() []string {
	keys := make([]string, 0, len(h.backends))
	for k := range h.backends {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (h *Handle) resolveChat(opts ChatOptions) (Backend, string, ChatOptions, error) {
	full := opts.Model
	if full == "" {
		full = h.defaultChatModel
	}
	if full == "" {
		return nil, "", opts, &Error{Kind: KindInvalidRequest, Message: "no model specified and no default_chat_model configured"}
	}
	prefix, name, ok := SplitModel(full)
	if !ok {
		return nil, "", opts, &Error{Kind: KindInvalidRequest, Message: "model '" + full + "' must include a provider prefix (e.g. 'openai/gpt-4o')"}
	}
	backend, ok := h.backends[prefix]
	if !ok {
		return nil, "", opts, &Error{
			Kind:    KindUnknownProvider,
			Message: "unknown model prefix '" + prefix + "'; known: " + strings.Join(h.BackendKeys(), ", "),
		}
	}
	resolved := opts
	if resolved.Temperature <= 0 {
		resolved.Temperature = h.defaultTemperature
	}
	if resolved.MaxTokens <= 0 {
		resolved.MaxTokens = h.defaultMaxTokens
	}
	return backend, name, resolved, nil
}

func (h *Handle) resolveEmbed(opts EmbedOptions) (Backend, string, error) {
	full := opts.Model
	if full == "" {
		full = h.defaultEmbeddingModel
	}
	if full == "" {
		return nil, "", &Error{Kind: KindInvalidRequest, Message: "no model specified and no default_embedding_model configured"}
	}
	prefix, name, ok := SplitModel(full)
	if !ok {
		return nil, "", &Error{Kind: KindInvalidRequest, Message: "model '" + full + "' must include a provider prefix"}
	}
	backend, ok := h.backends[prefix]
	if !ok {
		return nil, "", &Error{
			Kind:    KindUnknownProvider,
			Message: "unknown model prefix '" + prefix + "'; known: " + strings.Join(h.BackendKeys(), ", "),
		}
	}
	return backend, name, nil
}

// Chat is a non-streaming completion. ctx is honored (timeout/cancel).
func (h *Handle) Chat(ctx context.Context, messages []ChatMessage, opts ChatOptions) (*ChatResponse, error) {
	backend, model, resolved, err := h.resolveChat(opts)
	if err != nil {
		return nil, err
	}
	return backend.Chat(ctx, model, messages, resolved)
}

// ChatStream returns a channel of stream chunks. The channel always
// closes; callers should range over it.
func (h *Handle) ChatStream(ctx context.Context, messages []ChatMessage, opts ChatOptions) (<-chan StreamChunk, error) {
	backend, model, resolved, err := h.resolveChat(opts)
	if err != nil {
		return nil, err
	}
	return backend.ChatStream(ctx, model, messages, resolved)
}

// Embed runs the configured embedding backend over texts.
func (h *Handle) Embed(ctx context.Context, texts []string, opts EmbedOptions) (*EmbedResponse, error) {
	if len(texts) == 0 {
		return nil, &Error{Kind: KindInvalidRequest, Message: "texts must be non-empty"}
	}
	backend, model, err := h.resolveEmbed(opts)
	if err != nil {
		return nil, err
	}
	return backend.Embed(ctx, model, texts, opts)
}

// Keep errors as a tiny re-export anchor so consumers can import only
// this package.
var (
	_ = errors.New
)
