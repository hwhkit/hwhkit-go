// Package apiresponse defines the canonical JSON envelope used by every
// hwhkit service when returning a result to an HTTP/RPC caller.
//
// The shape is intentionally identical to the Rust (hwhkit-rs) and
// Python (hwhkit-py) implementations so that a client written against
// one runtime can decode responses from any of them:
//
//	{
//	    "code":     <int>,
//	    "message":  <string>,
//	    "data":     <T | null>,
//	    "trace_id": <string | "">
//	}
//
// Conventions:
//
//   - Code == 0 means success. Any non-zero value is an error.
//   - For errors we follow the hwhkit-py 6-digit error-code scheme
//     (e.g. 100001 = generic client error, 200001 = generic server
//     error). The actual codes live in higher-level packages; this
//     package only carries the integer.
//   - TraceID is the correlation id propagated from the request
//     (W3C traceparent or service-generated). It is omitted from
//     wire when empty.
//   - Data is a pointer so that the zero value of T does not get
//     conflated with "no payload". On the wire, a nil Data is encoded
//     as an absent field (omitempty), matching hwhkit-rs's
//     `Option<T>` behaviour.
package apiresponse

// CodeOK is the canonical success code. It matches hwhkit-rs's
// `ApiResponse::ok` and hwhkit-py's `ApiResponse.ok()` constructors.
const CodeOK = 0

// MessageOK is the default success message. Callers may override it
// (e.g. for localisation) by constructing the struct directly.
const MessageOK = "ok"

// ApiResponse is the generic envelope wrapping a response payload of
// type T. The pointer to T lets us distinguish "no data" (nil) from
// "zero-value data" (e.g. an empty struct).
//
// Example:
//
//	type User struct{ Name string `json:"name"` }
//	resp := apiresponse.OK(User{Name: "alice"})
//	// json: {"code":0,"message":"ok","data":{"name":"alice"}}
type ApiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *T     `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// OK builds a successful response carrying the given payload.
//
// Example:
//
//	resp := apiresponse.OK(42)
//	// resp.Code == 0, resp.Message == "ok", *resp.Data == 42
func OK[T any](data T) ApiResponse[T] {
	return ApiResponse[T]{
		Code:    CodeOK,
		Message: MessageOK,
		Data:    &data,
	}
}

// Err builds an error response. The data field is left nil so callers
// don't accidentally leak partial results on the error path.
//
// `code` should be a 6-digit error code consistent with the
// hwhkit-py taxonomy (1xxxxx = client, 2xxxxx = server, etc.).
//
// Example:
//
//	resp := apiresponse.Err[any](100404, "user not found")
//	// resp.Code == 100404, resp.Data == nil
func Err[T any](code int, message string) ApiResponse[T] {
	return ApiResponse[T]{
		Code:    code,
		Message: message,
	}
}

// WithTraceID returns a copy of r with TraceID set. It is a no-op
// receiver-style helper so that handlers can chain:
//
//	return apiresponse.OK(payload).WithTraceID(traceID)
func (r ApiResponse[T]) WithTraceID(traceID string) ApiResponse[T] {
	r.TraceID = traceID
	return r
}
