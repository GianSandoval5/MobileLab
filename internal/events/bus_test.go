package events

import (
	"context"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestBusPublishesVersionedEvents(t *testing.T) {
	bus := NewBus(2)
	defer bus.Close()
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()
	stream, cancel, err := bus.Subscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	event := domain.Event{Type: domain.EventRequestRecorded, Version: 1, Timestamp: time.Now(), Payload: "request"}
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	select {
	case received := <-stream:
		if received.Type != event.Type || received.Payload != "request" {
			t.Fatalf("unexpected event: %#v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("event not delivered")
	}
}

func TestBusDoesNotBlockOnSlowSubscriber(t *testing.T) {
	bus := NewBus(1)
	defer bus.Close()
	stream, cancel, err := bus.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	for index := 1; index <= 3; index++ {
		if err := bus.Publish(context.Background(), domain.Event{Type: domain.EventStateChanged, Version: 1, Payload: index}); err != nil {
			t.Fatal(err)
		}
	}
	if received := <-stream; received.Payload != 3 {
		t.Fatalf("slow subscriber did not retain newest event: %#v", received)
	}
}

func TestBusRejectsInvalidEventsAndClosesSubscribers(t *testing.T) {
	bus := NewBus(1)
	stream, _, err := bus.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), domain.Event{}); err == nil {
		t.Fatal("expected invalid event error")
	}
	bus.Close()
	if _, open := <-stream; open {
		t.Fatal("subscriber remained open after bus close")
	}
}
