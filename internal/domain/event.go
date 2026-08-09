package domain

import (
	"context"
	"time"
)

type EventType string

const (
	EventSnapshot          EventType = "environment.snapshot"
	EventRequestRecorded   EventType = "request.recorded"
	EventStateChanged      EventType = "environment.state_changed"
	EventScenarioCompleted EventType = "scenario.completed"
)

type Event struct {
	Type      EventType `json:"type"`
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

type EventPublisher interface {
	Publish(context.Context, Event) error
}
