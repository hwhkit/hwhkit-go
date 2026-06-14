// Package llm provides a multi-provider LLM integration (chat + embeddings)
// for the hwhkit-go bootstrap pipeline. Backends are selected per call
// from the model prefix:
//
//	anthropic/claude-3-5-sonnet-20241022 -> Anthropic Messages API
//	openai/gpt-4o-mini                   -> OpenAI Chat Completions
//	deepseek/deepseek-chat               -> OpenAI-compatible host
//	moonshot/moonshot-v1-128k            -> OpenAI-compatible host
//	ollama/llama3.1:8b                   -> OpenAI-compatible host
//
// API keys come from cfg.Integrations.LLM.Providers.* and are NOT read
// from environment variables by the framework. Keep secrets in your
// config layer.
package llm

import "encoding/json"

// Role mirrors OpenAI's role field. Backends translate per their own
// vocabulary (Anthropic merges System into the top-level system field).
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage is one turn in a conversation. Multi-part content is not
// modeled at this layer (text only for v1). See README for the roadmap.
type ChatMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	// Name is optional and rare (mostly OpenAI tool messages).
	Name string `json:"name,omitempty"`
}

// System builds a system message.
func System(content string) ChatMessage { return ChatMessage{Role: RoleSystem, Content: content} }

// User builds a user message.
func User(content string) ChatMessage { return ChatMessage{Role: RoleUser, Content: content} }

// Assistant builds an assistant message.
func Assistant(content string) ChatMessage {
	return ChatMessage{Role: RoleAssistant, Content: content}
}

// ChatOptions controls a single chat call. Zero values are safe on every
// backend ("use the handle's defaults").
type ChatOptions struct {
	// Model is a prefix-qualified identifier (e.g. "openai/gpt-4o-mini").
	// Empty means "use Handle.DefaultChatModel".
	Model string
	// Temperature in [0.0, 2.0]. Zero (the Go default) means "use
	// Handle.DefaultTemperature".
	Temperature float32
	// MaxTokens caps the output. Zero means "use Handle.DefaultMaxTokens".
	MaxTokens int
	// Stop sequences. May be nil.
	Stop []string
}

// ChatResponse is the non-streaming chat return.
type ChatResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage"`
	// Raw is the verbatim backend response body.
	Raw json.RawMessage `json:"raw"`
}

// Usage reports token counts. Zero means "not reported by the backend".
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Total returns input + output tokens.
func (u Usage) Total() int { return u.InputTokens + u.OutputTokens }

// StreamChunkKind discriminates between delta and terminal frames.
type StreamChunkKind string

const (
	ChunkText StreamChunkKind = "text_delta"
	ChunkDone StreamChunkKind = "done"
)

// StreamChunk is one frame of a streaming chat response.
type StreamChunk struct {
	Kind         StreamChunkKind `json:"kind"`
	Text         string          `json:"text,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Usage        *Usage          `json:"usage,omitempty"`
	// Err, when non-nil, is a transport / decode failure; the consumer
	// should stop reading.
	Err error `json:"-"`
}

// EmbedOptions controls an Embed() call.
type EmbedOptions struct {
	// Model is a prefix-qualified identifier
	// (e.g. "openai/text-embedding-3-small"). Empty means
	// "use Handle.DefaultEmbeddingModel".
	Model string
}

// EmbedResponse holds one vector per input text plus aggregate usage.
type EmbedResponse struct {
	Vectors [][]float32 `json:"vectors"`
	Model   string      `json:"model"`
	Usage   Usage       `json:"usage"`
}
