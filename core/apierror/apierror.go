// Package apierror implements RFC 7807 application/problem+json handler errors.
package apierror

import (
	"encoding/json"
	"net/http"
)

const ContentType = "application/problem+json"

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type ApiError struct {
	Status int          `json:"status"`
	Type   string       `json:"type,omitempty"`
	Title  string       `json:"title"`
	Detail string       `json:"detail,omitempty"`
	Code   string       `json:"code,omitempty"`
	Fields []FieldError `json:"fields,omitempty"`
}

func (e *ApiError) Error() string {
	if e.Detail != "" {
		return e.Title + ": " + e.Detail
	}
	return e.Title
}

func (e *ApiError) WriteJSON(w http.ResponseWriter) {
	if e.Status == 0 {
		e.Status = http.StatusInternalServerError
	}
	if e.Title == "" {
		e.Title = http.StatusText(e.Status)
	}
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(e)
}

func New(status int, title, detail string) *ApiError {
	return &ApiError{Status: status, Title: title, Detail: detail}
}

func NotFound(detail string) *ApiError {
	return &ApiError{Status: http.StatusNotFound, Title: "Not Found", Detail: detail}
}

func Unauthorized(detail string) *ApiError {
	return &ApiError{Status: http.StatusUnauthorized, Title: "Unauthorized", Detail: detail}
}

func Forbidden(detail string) *ApiError {
	return &ApiError{Status: http.StatusForbidden, Title: "Forbidden", Detail: detail}
}

func BadRequest(detail string) *ApiError {
	return &ApiError{Status: http.StatusBadRequest, Title: "Bad Request", Detail: detail}
}

func Conflict(detail string) *ApiError {
	return &ApiError{Status: http.StatusConflict, Title: "Conflict", Detail: detail}
}

func Validation(fields []FieldError) *ApiError {
	return &ApiError{
		Status: http.StatusUnprocessableEntity,
		Title:  "Validation Failed",
		Fields: fields,
	}
}

func TooManyRequests(detail string) *ApiError {
	return &ApiError{Status: http.StatusTooManyRequests, Title: "Too Many Requests", Detail: detail}
}

func Internal(detail string) *ApiError {
	return &ApiError{Status: http.StatusInternalServerError, Title: "Internal Server Error", Detail: detail}
}
