package recording

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestServiceRecordsOneOrderedSession(t *testing.T) {
	service := NewService()
	now := time.Date(2026, 8, 9, 21, 0, 0, 0, time.UTC)
	service.SetClock(func() time.Time { return now })
	started, err := service.StartRecording(context.Background(), "login", domain.EnvironmentState{LatencyMS: 20})
	if err != nil || started.Name != "login" {
		t.Fatalf("start: %#v %v", started, err)
	}
	if _, err := service.StartRecording(context.Background(), "other", domain.EnvironmentState{}); err == nil {
		t.Fatal("expected active-session error")
	}

	var wait sync.WaitGroup
	for index := 0; index < 10; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if err := service.RecordCapture(context.Background(), domain.CaptureEvent{
				Kind: domain.CaptureDeepLink, DeepLink: &domain.DeepLinkCapture{URL: "app://login"},
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wait.Wait()
	stopped, err := service.StopRecording(context.Background())
	if err != nil || len(stopped.Events) != 10 {
		t.Fatalf("stop: events=%d err=%v", len(stopped.Events), err)
	}
	for index, event := range stopped.Events {
		if event.Sequence != int64(index+1) || event.Timestamp != now {
			t.Fatalf("event %d: %#v", index, event)
		}
	}
}

func TestServiceValidatesNamesAndEvents(t *testing.T) {
	service := NewService()
	if _, err := service.StartRecording(context.Background(), "../login", domain.EnvironmentState{}); err == nil {
		t.Fatal("expected unsafe name rejection")
	}
	if err := service.RecordCapture(context.Background(), domain.CaptureEvent{Kind: domain.CaptureDeepLink}); err == nil {
		t.Fatal("expected invalid event rejection")
	}
}
