package storage

import (
	"context"
	"sync"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type MemoryRequests struct {
	mu      sync.RWMutex
	limit   int
	records []domain.RequestRecord
}

func NewMemoryRequests(limit int) *MemoryRequests {
	if limit < 1 {
		limit = 500
	}
	return &MemoryRequests{limit: limit}
}

func (m *MemoryRequests) Append(_ context.Context, record domain.RequestRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.records) == m.limit {
		copy(m.records, m.records[1:])
		m.records[len(m.records)-1] = record
		return nil
	}
	m.records = append(m.records, record)
	return nil
}

func (m *MemoryRequests) Recent(_ context.Context, limit int) ([]domain.RequestRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > len(m.records) {
		limit = len(m.records)
	}
	start := len(m.records) - limit
	result := make([]domain.RequestRecord, limit)
	copy(result, m.records[start:])
	return result, nil
}
