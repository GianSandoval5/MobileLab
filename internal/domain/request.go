package domain

import (
	"context"
	"time"
)

type RequestRecord struct {
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Query           map[string][]string `json:"query,omitempty"`
	Headers         map[string][]string `json:"headers,omitempty"`
	Body            any                 `json:"body,omitempty"`
	Status          int                 `json:"status"`
	ResponseHeaders map[string][]string `json:"response_headers,omitempty"`
	ResponseBody    any                 `json:"response_body,omitempty"`
	DurationMS      int64               `json:"duration_ms"`
	Timestamp       time.Time           `json:"timestamp"`
}

type RequestRepository interface {
	Append(context.Context, RequestRecord) error
	Recent(context.Context, int) ([]RequestRecord, error)
}
