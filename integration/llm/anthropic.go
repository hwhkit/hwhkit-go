package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicBackend speaks the Anthropic Messages API (POST /v1/messages).
type AnthropicBackend struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	opTimeout time.Duration
}

// NewAnthropic builds a backend. baseURL may be empty (defaults to the
// public host).
func NewAnthropic(apiKey, baseURL string, opTimeout time.Duration) *AnthropicBackend {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if opTimeout <= 0 {
		opTimeout = 30 * time.Second
	}
	return &AnthropicBackend{
		client:    &http.Client{Timeout: opTimeout},
		apiKey:    apiKey,
		baseURL:   strings.TrimRight(baseURL, "/"),
		opTimeout: opTimeout,
	}
}

// Key returns the routing prefix.
func (b *AnthropicBackend) Key() string { return "anthropic" }

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicReq struct {
	Model         string         `json:"model"`
	Messages      []anthropicMsg `json:"messages"`
	System        string         `json:"system,omitempty"`
	Temperature   *float32       `json:"temperature,omitempty"`
	MaxTokens     int            `json:"max_tokens"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	Stream        bool           `json:"stream,omitempty"`
}

type anthropicResp struct {
	Model      string                   `json:"model"`
	Content    []map[string]interface{} `json:"content"`
	StopReason string                   `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// splitSystem peels system messages out of `messages` and merges them
// into the top-level system field per Anthropic's wire shape.
func splitSystemAnthropic(messages []ChatMessage) (string, []anthropicMsg) {
	var system string
	wire := make([]anthropicMsg, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			if system == "" {
				system = m.Content
			} else {
				system = system + "\n\n" + m.Content
			}
		case RoleAssistant:
			wire = append(wire, anthropicMsg{Role: "assistant", Content: m.Content})
		default:
			// User + Tool collapse to "user" since this API doesn't
			// model tool results as a distinct role at this layer.
			wire = append(wire, anthropicMsg{Role: "user", Content: m.Content})
		}
	}
	return system, wire
}

func buildAnthropicReq(model string, messages []ChatMessage, opts ChatOptions, stream bool) anthropicReq {
	system, wire := splitSystemAnthropic(messages)
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		// Anthropic requires max_tokens — default to a soft cap.
		maxTokens = 4096
	}
	var temp *float32
	if opts.Temperature > 0 {
		t := opts.Temperature
		temp = &t
	}
	return anthropicReq{
		Model:         model,
		Messages:      wire,
		System:        system,
		Temperature:   temp,
		MaxTokens:     maxTokens,
		StopSequences: opts.Stop,
		Stream:        stream,
	}
}

func (b *AnthropicBackend) buildRequest(ctx context.Context, body []byte, stream bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, &Error{Kind: KindInvalidRequest, Backend: b.Key(), Message: "build request", Cause: err}
	}
	req.Header.Set("x-api-key", b.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	if stream {
		req.Header.Set("accept", "text/event-stream")
	}
	return req, nil
}

// Chat implements Backend.Chat.
func (b *AnthropicBackend) Chat(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (*ChatResponse, error) {
	body, _ := json.Marshal(buildAnthropicReq(model, messages, opts, false))
	req, err := b.buildRequest(ctx, body, false)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError(b.Key(), err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &Error{Kind: KindBadStatus, Backend: b.Key(), Status: resp.StatusCode, Message: truncate(string(raw))}
	}
	var parsed anthropicResp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &Error{Kind: KindDecode, Backend: b.Key(), Message: err.Error()}
	}
	var content strings.Builder
	for _, block := range parsed.Content {
		if t, _ := block["type"].(string); t == "text" {
			if s, _ := block["text"].(string); s != "" {
				content.WriteString(s)
			}
		}
	}
	finish := parsed.StopReason
	if finish == "" {
		finish = "stop"
	}
	return &ChatResponse{
		Content:      content.String(),
		Model:        parsed.Model,
		FinishReason: finish,
		Usage:        Usage{InputTokens: parsed.Usage.InputTokens, OutputTokens: parsed.Usage.OutputTokens},
		Raw:          raw,
	}, nil
}

// ChatStream implements Backend.ChatStream.
func (b *AnthropicBackend) ChatStream(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (<-chan StreamChunk, error) {
	body, _ := json.Marshal(buildAnthropicReq(model, messages, opts, true))
	req, err := b.buildRequest(ctx, body, true)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError(b.Key(), err)
	}
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &Error{Kind: KindBadStatus, Backend: b.Key(), Status: resp.StatusCode, Message: truncate(string(raw))}
	}

	out := make(chan StreamChunk, 16)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		pumpAnthropicSSE(resp.Body, out)
	}()
	return out, nil
}

func pumpAnthropicSSE(body io.Reader, out chan<- StreamChunk) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			return
		}
		var evt map[string]json.RawMessage
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			continue
		}
		var ty string
		_ = json.Unmarshal(evt["type"], &ty)
		switch ty {
		case "content_block_delta":
			var delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			_ = json.Unmarshal(evt["delta"], &delta)
			if delta.Text != "" {
				out <- StreamChunk{Kind: ChunkText, Text: delta.Text}
			}
		case "message_delta":
			var delta struct {
				StopReason string `json:"stop_reason"`
			}
			_ = json.Unmarshal(evt["delta"], &delta)
			if delta.StopReason != "" {
				out <- StreamChunk{Kind: ChunkDone, FinishReason: delta.StopReason}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		out <- StreamChunk{Err: &Error{Kind: KindTransport, Backend: "anthropic", Message: err.Error()}}
	}
}

// Embed is not supported on Anthropic.
func (b *AnthropicBackend) Embed(_ context.Context, _ string, _ []string, _ EmbedOptions) (*EmbedResponse, error) {
	return nil, &Error{Kind: KindInvalidRequest, Backend: b.Key(), Message: "anthropic backend does not support embeddings"}
}

func classifyHTTPError(backend string, err error) error {
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "Client.Timeout") {
		return &Error{Kind: KindTimeout, Backend: backend, Message: msg, Cause: err}
	}
	return &Error{Kind: KindTransport, Backend: backend, Message: msg, Cause: err}
}

// Suppress unused import warning when downstream code doesn't reach it
// in some builds.
var _ = fmt.Sprintf
