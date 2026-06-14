package apiresponse

import (
	"encoding/json"
	"testing"
)

type payload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestOK(t *testing.T) {
	resp := OK(payload{Name: "alice", Count: 3})
	if resp.Code != CodeOK {
		t.Fatalf("OK code = %d, want %d", resp.Code, CodeOK)
	}
	if resp.Message != MessageOK {
		t.Fatalf("OK message = %q, want %q", resp.Message, MessageOK)
	}
	if resp.Data == nil {
		t.Fatal("OK data is nil")
	}
	if resp.Data.Name != "alice" || resp.Data.Count != 3 {
		t.Fatalf("OK data = %+v", resp.Data)
	}
	if resp.TraceID != "" {
		t.Fatalf("OK trace_id should be empty by default, got %q", resp.TraceID)
	}
}

func TestErr(t *testing.T) {
	resp := Err[payload](100404, "not found")
	if resp.Code != 100404 {
		t.Fatalf("Err code = %d, want 100404", resp.Code)
	}
	if resp.Message != "not found" {
		t.Fatalf("Err message = %q", resp.Message)
	}
	if resp.Data != nil {
		t.Fatalf("Err data should be nil, got %+v", resp.Data)
	}
}

func TestWithTraceID(t *testing.T) {
	const tid = "01HF000000000000000000"
	resp := OK(payload{Name: "bob"}).WithTraceID(tid)
	if resp.TraceID != tid {
		t.Fatalf("WithTraceID = %q, want %q", resp.TraceID, tid)
	}
	// Should not mutate the original (value receiver).
	base := OK(payload{Name: "bob"})
	_ = base.WithTraceID(tid)
	if base.TraceID != "" {
		t.Fatalf("WithTraceID mutated receiver: %q", base.TraceID)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Success envelope with trace id.
	src := OK(payload{Name: "carol", Count: 7}).WithTraceID("trace-xyz")
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"code":0,"message":"ok","data":{"name":"carol","count":7},"trace_id":"trace-xyz"}`
	if string(raw) != want {
		t.Fatalf("encoded = %s\nwant     = %s", raw, want)
	}

	var got ApiResponse[payload]
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Code != 0 || got.Message != "ok" || got.TraceID != "trace-xyz" {
		t.Fatalf("decoded scalar fields = %+v", got)
	}
	if got.Data == nil || got.Data.Name != "carol" || got.Data.Count != 7 {
		t.Fatalf("decoded data = %+v", got.Data)
	}

	// Error envelope: data omitted, trace_id omitted.
	errResp := Err[payload](200500, "boom")
	raw, err = json.Marshal(errResp)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	if string(raw) != `{"code":200500,"message":"boom"}` {
		t.Fatalf("error envelope = %s", raw)
	}
}
