package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type Bus struct {
	mu          sync.RWMutex
	buffer      int
	nextID      uint64
	subscribers map[uint64]chan domain.Event
	closed      bool
}

func NewBus(buffer int) *Bus {
	if buffer < 1 {
		buffer = 64
	}
	return &Bus{buffer: buffer, subscribers: make(map[uint64]chan domain.Event)}
}

func (b *Bus) Publish(ctx context.Context, event domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.Type == "" || event.Version < 1 {
		return fmt.Errorf("event type and positive version are required")
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return fmt.Errorf("event bus is closed")
	}
	for _, subscriber := range b.subscribers {
		select {
		case subscriber <- event:
		default:
			// Preserve current state for slow dashboard consumers by evicting
			// their oldest buffered event instead of blocking the sandbox.
			select {
			case <-subscriber:
			default:
			}
			select {
			case subscriber <- event:
			default:
			}
		}
	}
	return nil
}

func (b *Bus) Subscribe(ctx context.Context) (<-chan domain.Event, func(), error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, nil, fmt.Errorf("event bus is closed")
	}
	id := b.nextID
	b.nextID++
	channel := make(chan domain.Event, b.buffer)
	b.subscribers[id] = channel
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			if existing, found := b.subscribers[id]; found {
				delete(b.subscribers, id)
				close(existing)
			}
			b.mu.Unlock()
		})
	}
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return channel, cancel, nil
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for id, subscriber := range b.subscribers {
		delete(b.subscribers, id)
		close(subscriber)
	}
}

var _ domain.EventPublisher = (*Bus)(nil)
