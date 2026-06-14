package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSplitModel(t *testing.T) {
	cases := []struct {
		in            string
		wantOK        bool
		wantP, wantN  string
	}{
		{"openai/gpt-4o", true, "openai", "gpt-4o"},
		{"anthropic/claude-3-5-sonnet", true, "anthropic", "claude-3-5-sonnet"},
		{"no-prefix", false, "", ""},
		{"/empty-prefix", false, "", ""},
		{"empty-name/", false, "", ""},
		{"", false, "", ""},
	}
	for _, c := range cases {
		p, n, ok := SplitModel(c.in)
		if ok != c.wantOK || p != c.wantP || n != c.wantN {
			t.Errorf("SplitModel(%q) = (%q,%q,%v), want (%q,%q,%v)",
				c.in, p, n, ok, c.wantP, c.wantN, c.wantOK)
		}
	}
}

func TestHandleResolveErrors(t *testing.T) {
	h := NewHandle(nil, "", "", 0.7, 0, time.Second)

	// no default and no opts.Model
	if _, err := h.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{}); !IsKind(err, KindInvalidRequest) {
		t.Errorf("expected InvalidRequest, got %v", err)
	}

	// missing prefix
	h2 := NewHandle(nil, "claude-3-5-sonnet", "", 0.7, 0, time.Second)
	if _, err := h2.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{}); !IsKind(err, KindInvalidRequest) {
		t.Errorf("expected InvalidRequest (no prefix), got %v", err)
	}

	// unknown prefix
	h3 := NewHandle(nil, "anthropic/claude-3-5-sonnet", "", 0.7, 0, time.Second)
	if _, err := h3.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{}); !IsKind(err, KindUnknownProvider) {
		t.Errorf("expected UnknownProvider, got %v", err)
	}
}

func TestOpenAIChatHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
            "model": "gpt-4o-mini",
            "choices": [{"message": {"content": "hi from gpt"}, "finish_reason": "stop"}],
            "usage": {"prompt_tokens": 9, "completion_tokens": 3}
        }`))
	}))
	defer srv.Close()

	backend := NewOpenAICompat("openai", "k", srv.URL, 5*time.Second)
	h := NewHandle([]Backend{backend}, "openai/gpt-4o-mini", "", 0.7, 0, 5*time.Second)

	resp, err := h.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hi from gpt" {
		t.Errorf("Content = %q, want %q", resp.Content, "hi from gpt")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", resp.FinishReason)
	}
	if resp.Usage.InputTokens != 9 || resp.Usage.OutputTokens != 3 {
		t.Errorf("Usage = %+v, want 9/3", resp.Usage)
	}
}

func TestOpenAIChatBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"bad key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()
	backend := NewOpenAICompat("openai", "k", srv.URL, 5*time.Second)
	h := NewHandle([]Backend{backend}, "openai/gpt-4o-mini", "", 0.7, 0, 5*time.Second)
	_, err := h.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{})
	if !IsKind(err, KindBadStatus) {
		t.Fatalf("expected BadStatus, got %v", err)
	}
}

func TestOpenAIChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"hello"},"finish_reason":null}]}` + "\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":" world"},"finish_reason":null}]}` + "\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{},"finish_reason":"stop"}]}` + "\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte(`data: [DONE]` + "\n\n"))
	}))
	defer srv.Close()

	backend := NewOpenAICompat("openai", "k", srv.URL, 5*time.Second)
	h := NewHandle([]Backend{backend}, "openai/gpt-4o-mini", "", 0.7, 0, 5*time.Second)
	ch, err := h.ChatStream(context.Background(), []ChatMessage{User("hi")}, ChatOptions{})
	if err != nil {
		t.Fatalf("ChatStream err: %v", err)
	}
	var buf strings.Builder
	var finish string
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatalf("chunk err: %v", chunk.Err)
		}
		switch chunk.Kind {
		case ChunkText:
			buf.WriteString(chunk.Text)
		case ChunkDone:
			finish = chunk.FinishReason
		}
	}
	if buf.String() != "hello world" {
		t.Errorf("text = %q, want %q", buf.String(), "hello world")
	}
	if finish != "stop" {
		t.Errorf("finish = %q, want stop", finish)
	}
}

func TestOpenAIEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
            "model": "text-embedding-3-small",
            "data": [
                {"embedding": [0.1, 0.2, 0.3]},
                {"embedding": [0.4, 0.5, 0.6]}
            ],
            "usage": {"prompt_tokens": 8, "completion_tokens": 0}
        }`))
	}))
	defer srv.Close()
	backend := NewOpenAICompat("openai", "k", srv.URL, 5*time.Second)
	h := NewHandle([]Backend{backend}, "", "openai/text-embedding-3-small", 0.7, 0, 5*time.Second)
	resp, err := h.Embed(context.Background(), []string{"a", "b"}, EmbedOptions{})
	if err != nil {
		t.Fatalf("Embed err: %v", err)
	}
	if len(resp.Vectors) != 2 || resp.Vectors[0][0] != 0.1 {
		t.Errorf("Vectors = %+v", resp.Vectors)
	}
}

func TestAnthropicChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
            "model": "claude-3-5-sonnet-20241022",
            "content": [
                {"type": "text", "text": "Hello "},
                {"type": "text", "text": "world."}
            ],
            "stop_reason": "end_turn",
            "usage": {"input_tokens": 12, "output_tokens": 4}
        }`))
	}))
	defer srv.Close()
	backend := NewAnthropic("k", srv.URL, 5*time.Second)
	h := NewHandle([]Backend{backend}, "anthropic/claude-3-5-sonnet-20241022", "", 0.7, 0, 5*time.Second)
	resp, err := h.Chat(context.Background(), []ChatMessage{User("hi")}, ChatOptions{})
	if err != nil {
		t.Fatalf("Chat err: %v", err)
	}
	if resp.Content != "Hello world." {
		t.Errorf("Content = %q", resp.Content)
	}
	if resp.FinishReason != "end_turn" {
		t.Errorf("FinishReason = %q", resp.FinishReason)
	}
	if resp.Usage.InputTokens != 12 {
		t.Errorf("Usage.InputTokens = %d", resp.Usage.InputTokens)
	}
}
