package sandbox

import "sync"

type RuntimeState struct {
	mu          sync.RWMutex
	latencyMS   int
	forcedError int
	authExpired bool
}

type StateSnapshot struct {
	LatencyMS   int  `json:"latency_ms"`
	ForcedError int  `json:"forced_error,omitempty"`
	AuthExpired bool `json:"auth_expired"`
}

func NewRuntimeState(latencyMS int) *RuntimeState {
	return &RuntimeState{latencyMS: latencyMS}
}

func (s *RuntimeState) Snapshot() StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return StateSnapshot{LatencyMS: s.latencyMS, ForcedError: s.forcedError, AuthExpired: s.authExpired}
}

func (s *RuntimeState) SetLatency(milliseconds int) {
	s.mu.Lock()
	s.latencyMS = milliseconds
	s.mu.Unlock()
}

func (s *RuntimeState) SetError(status int) {
	s.mu.Lock()
	s.forcedError = status
	s.mu.Unlock()
}

func (s *RuntimeState) SetAuthExpired(expired bool) {
	s.mu.Lock()
	s.authExpired = expired
	s.mu.Unlock()
}

func (s *RuntimeState) Reset() {
	s.mu.Lock()
	s.latencyMS = 0
	s.forcedError = 0
	s.authExpired = false
	s.mu.Unlock()
}
