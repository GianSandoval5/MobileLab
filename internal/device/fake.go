package device

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type FakeAdapter struct {
	mu           sync.Mutex
	PlatformName string
	Devices      []domain.Device
	Operations   []string
	Error        error
}

func (f *FakeAdapter) Platform() string {
	if f.PlatformName != "" {
		return f.PlatformName
	}
	return "fake"
}

func (f *FakeAdapter) Detect(context.Context) ([]domain.Device, error) {
	return append([]domain.Device(nil), f.Devices...), f.Error
}

func (f *FakeAdapter) LaunchApp(_ context.Context, deviceID, appID string) error {
	return f.record("launch:" + deviceID + ":" + appID)
}

func (f *FakeAdapter) StopApp(_ context.Context, deviceID, appID string) error {
	return f.record("stop:" + deviceID + ":" + appID)
}

func (f *FakeAdapter) ClearApp(_ context.Context, deviceID, appID string) error {
	return f.record("clear:" + deviceID + ":" + appID)
}

func (f *FakeAdapter) BootDevice(_ context.Context, deviceID string) error {
	return f.record("boot:" + deviceID)
}

func (f *FakeAdapter) OpenDeepLink(_ context.Context, deviceID, value string) error {
	return f.record("deeplink:" + deviceID + ":" + value)
}

func (f *FakeAdapter) SetLocation(_ context.Context, deviceID string, location domain.Location) error {
	return f.record("location:" + deviceID + ":" + strconvLocation(location))
}

func (f *FakeAdapter) SetNetworkCondition(_ context.Context, deviceID string, condition domain.NetworkCondition) error {
	return f.record("network:" + deviceID + ":" + string(condition))
}

func (f *FakeAdapter) record(operation string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Operations = append(f.Operations, operation)
	return f.Error
}

func strconvLocation(location domain.Location) string {
	return fmt.Sprintf("%g,%g", location.Latitude, location.Longitude)
}
