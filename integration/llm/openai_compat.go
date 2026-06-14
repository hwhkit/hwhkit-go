package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatBackend speaks the OpenAI Chat Completions wire shape.
// One instance is bound to one host (openai.com / deepseek / moonshot /
// ollama / a LiteLLM proxy / ...).
type OpenAICompatBackend struct {
	key       string
	client    *http.Client
	apiKey    string
	baseURL   string
	opTimeout time.Duration
}

// NewOpenAICompat builds a backend. `key` is the routing prefix
// ("openai", "deepseek", "moonshot", "ollama"). baseURL may be empty —
// known prefixes get a sensible default.
func NewOpenAICompat(key, apiKey, baseURL string, opTimeout time.Duration) *OpenAICompatBackend {
	if baseURL == "" {
		baseURL = defaultBase(key)
	}
	if opTimeout <= 0 {
		opTimeout = 30 * time.Second
	}
	return &OpenAICompatBackend{
		key:       key,
		client:    &http.Client{Timeout: opTimeout},
		apiKey:    apiKey,
		baseURL:   strings.TrimRight(baseURL, "/"),
		opTimeout: opTimeout,
	}
}

func defaultBase(key string) string {
	switch key {
	case "openai":
		return "https://api.openai.com"
	case "deepseek":
		return "https://api.deepseek.com"
	case "moonshot":
		return "https://api.moonshot.cn"
	case "ollama":
		return "http://localhost:11434"
	}
	return "https://api.openai.com"
}

// Key returns the routing prefix.
func (b *OpenAICompatBackend) Key() string { return b.key }

type openAIMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type openAIReq struct {
	Model       string      `json:"model"`
	Messages    []openAIMsg `json:"messages"`
	Temperature *float32    `json:"temperature,omitempty"`
	MaxTokens   *int        `json:"max_tokens,omitempty"`
	Stop        []string    `json:"stop,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
}

type openAIResp struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func buildOpenAIReq(model string, messages []ChatMessage, opts ChatOptions, stream bool) openAIReq {
	wire := make([]openAIMsg, 0, len(messages))
	for _, m := range messages {
		wire = append(wire, openAIMsg{Role: string(m.Role), Content: m.Content, Name: m.Name})
	}
	var temp *float32
	if opts.Temperature > 0 {
		t := opts.Temperature
		temp = &t
	}
	var mt *int
	if opts.MaxTokens > 0 {
		v := opts.MaxTokens
		mt = &v
	}
	return openAIReq{
		Model:       model,
		Messages:    wire,
		Temperature: temp,
		MaxTokens:   mt,
		Stop:        opts.Stop,
		Stream:      stream,
	}
}

func (b *OpenAICompatBackend) buildRequest(ctx context.Context, urlPath string, body []byte, stream bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+urlPath, bytes.NewReader(body))
	if err != nil {
		return nil, &Error{Kind: KindInvalidRequest, Backend: b.key, Message: "build request", Cause: err}
	}
	if b.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.apiKey)
	}
	req.Header.Set("content-type", "application/json")
	if stream {
		req.Header.Set("accept", "text/event-stream")
	}
	return req, nil
}

// Chat implements Backend.Chat.
func (b *OpenAICompatBackend) Chat(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (*ChatResponse, error) {
	body, _ := json.Marshal(buildOpenAIReq(model, messages, opts, false))
	req, err := b.buildRequest(ctx, "/v1/chat/completions", body, false)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError(b.key, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &Error{Kind: KindBadStatus, Backend: b.key, Status: resp.StatusCode, Message: truncate(string(raw))}
	}
	var parsed openAIResp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &Error{Kind: KindDecode, Backend: b.key, Message: err.Error()}
	}
	if len(parsed.Choices) == 0 {
		return nil, &Error{Kind: KindDecode, Backend: b.key, Message: "response has no choices"}
	}
	finish := parsed.Choices[0].FinishReason
	if finish == "" {
		finish = "stop"
	}
	return &ChatResponse{
		Content:      parsed.Choices[0].Message.Content,
		Model:        parsed.Model,
		FinishReason: finish,
		Usage:        Usage{InputTokens: parsed.Usage.PromptTokens, OutputTokens: parsed.Usage.CompletionTokens},
		Raw:          raw,
	}, nil
}

// ChatStream implements Backend.ChatStream.
func (b *OpenAICompatBackend) ChatStream(ctx context.Context, model string, messages []ChatMessage, opts ChatOptions) (<-chan StreamChunk, error) {
	body, _ := json.Marshal(buildOpenAIReq(model, messages, opts, true))
	req, err := b.buildRequest(ctx, "/v1/chat/completions", body, true)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError(b.key, err)
	}
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &Error{Kind: KindBadStatus, Backend: b.key, Status: resp.StatusCode, Message: truncate(string(raw))}
	}
	out := make(chan StreamChunk, 16)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		pumpOpenAISSE(resp.Body, out)
	}()
	return out, nil
}

func pumpOpenAISSE(body io.Reader, out chan<- StreamChunk) {
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
		var evt struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			continue
		}
		if len(evt.Choices) == 0 {
			continue
		}
		c := evt.Choices[0]
		if c.Delta.Content != "" {
			out <- StreamChunk{Kind: ChunkText, Text: c.Delta.Content}
		}
		if c.FinishReason != nil && *c.FinishReason != "" {
			out <- StreamChunk{Kind: ChunkDone, FinishReason: *c.FinishReason}
		}
	}
}

// Embed implements Backend.Embed.
func (b *OpenAICompatBackend) Embed(ctx context.Context, model string, texts []string, _ EmbedOptions) (*EmbedResponse, error) {
	type req struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	type item struct {
		Embedding []float32 `json:"embedding"`
	}
	type resp struct {
		Model string `json:"model"`
		Data  []item `json:"data"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	body, _ := json.Marshal(req{Model: model, Input: texts})
	httpReq, err := b.buildRequest(ctx, "/v1/embeddings", body, false)
	if err != nil {
		return nil, err
	}
	httpResp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, classifyHTTPError(b.key, err)
	}
	defer httpResp.Body.Close()
	raw, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode/100 != 2 {
		return nil, &Error{Kind: KindBadStatus, Backend: b.key, Status: httpResp.StatusCode, Message: truncate(string(raw))}
	}
	var parsed resp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &Error{Kind: KindDecode, Backend: b.key, Message: err.Error()}
	}
	vecs := make([][]float32, len(parsed.Data))
	for i, d := range parsed.Data {
		vecs[i] = d.Embedding
	}
	return &EmbedResponse{
		Vectors: vecs,
		Model:   parsed.Model,
		Usage:   Usage{InputTokens: parsed.Usage.PromptTokens, OutputTokens: parsed.Usage.CompletionTokens},
	}, nil
}
