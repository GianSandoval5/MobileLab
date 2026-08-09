package recording

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

var recordingNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type Service struct {
	mu     sync.RWMutex
	active *domain.Recording
	now    func() time.Time
}

func NewService() *Service {
	return &Service{now: time.Now}
}

func (service *Service) SetClock(now func() time.Time) {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.now = now
}

func (service *Service) StartRecording(_ context.Context, name string, initial domain.EnvironmentState) (domain.Recording, error) {
	name = strings.TrimSpace(name)
	if !recordingNamePattern.MatchString(name) {
		return domain.Recording{}, fmt.Errorf("recording name must contain 1-64 letters, digits, dots, underscores, or hyphens")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.active != nil {
		return domain.Recording{}, fmt.Errorf("recording %q is already active", service.active.Name)
	}
	recording := domain.Recording{Name: name, StartedAt: service.now().UTC(), InitialEnvironment: initial, Events: []domain.CaptureEvent{}}
	service.active = &recording
	return clone(recording), nil
}

func (service *Service) RecordCapture(_ context.Context, event domain.CaptureEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.active == nil {
		return nil
	}
	event.Sequence = int64(len(service.active.Events) + 1)
	if event.Timestamp.IsZero() {
		event.Timestamp = service.now().UTC()
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}
	service.active.Events = append(service.active.Events, event)
	return nil
}

func (service *Service) StopRecording(_ context.Context) (domain.Recording, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.active == nil {
		return domain.Recording{}, fmt.Errorf("no recording is active")
	}
	service.active.StoppedAt = service.now().UTC()
	result := clone(*service.active)
	service.active = nil
	return result, nil
}

func (service *Service) ActiveRecording(_ context.Context) (domain.Recording, bool) {
	service.mu.RLock()
	defer service.mu.RUnlock()
	if service.active == nil {
		return domain.Recording{}, false
	}
	return clone(*service.active), true
}

func clone(recording domain.Recording) domain.Recording {
	recording.Events = append([]domain.CaptureEvent(nil), recording.Events...)
	return recording
}
