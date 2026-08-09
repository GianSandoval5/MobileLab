package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/device"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestVersion(t *testing.T) {
	var output bytes.Buffer
	runner := New(&output, &output, t.TempDir())
	if err := runner.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), Version) {
		t.Fatalf("output does not contain version: %q", output.String())
	}
}

func TestUnknownCommandIsActionable(t *testing.T) {
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	err := runner.Run(context.Background(), []string{"wat"})
	if err == nil || !strings.Contains(err.Error(), "mobilelab help") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScenarioListParsesGeneratedScenarios(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "mobilelab", "scenarios")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "payment.yaml"), []byte("name: Payment failure\nexpect:\n  - response: {status: 500}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	runner := New(&output, &output, root)
	if err := runner.Run(context.Background(), []string{"scenario", "list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "payment") || !strings.Contains(output.String(), "Payment failure") {
		t.Fatalf("unexpected scenario list: %q", output.String())
	}
}

func TestDeviceClearUsesExplicitSelectedDevice(t *testing.T) {
	adapter := &device.FakeAdapter{
		PlatformName: "android",
		Devices: []domain.Device{{
			ID: "emulator-5554", Name: "Pixel", Platform: "android", State: "device",
			Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityClear: domain.CapabilityAvailable},
		}},
	}
	var output bytes.Buffer
	runner := New(&output, &output, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	err := runner.Run(context.Background(), []string{"device", "clear", "--platform", "android", "--device", "emulator-5554", "--app-id", "dev.mobilelab.app"})
	if err != nil {
		t.Fatal(err)
	}
	if len(adapter.Operations) != 1 || adapter.Operations[0] != "clear:emulator-5554:dev.mobilelab.app" {
		t.Fatalf("unexpected operations: %v", adapter.Operations)
	}
}

func TestDeepLinkSelectionDoesNotFallBackToAnotherDevice(t *testing.T) {
	android := &device.FakeAdapter{
		PlatformName: "android",
		Devices:      []domain.Device{{ID: "android-1", Name: "Android", Platform: "android", State: "device", Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityDeepLink: domain.CapabilityAvailable}}},
	}
	ios := &device.FakeAdapter{
		PlatformName: "ios",
		Devices:      []domain.Device{{ID: "ios-1", Name: "iPhone", Platform: "ios", State: "Booted", Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityDeepLink: domain.CapabilityAvailable}}},
	}
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{android, ios}
	err := runner.Run(context.Background(), []string{"deeplink", "open", "myapp://test", "--platform", "ios", "--device", "ios-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(android.Operations) != 0 || len(ios.Operations) != 1 || !strings.HasPrefix(ios.Operations[0], "deeplink:ios-1:") {
		t.Fatalf("selection fell back to wrong adapter: android=%v ios=%v", android.Operations, ios.Operations)
	}
}

func TestDeviceBootRequiresAdvertisedCapability(t *testing.T) {
	adapter := &device.FakeAdapter{
		PlatformName: "ios",
		Devices:      []domain.Device{{ID: "ios-1", Name: "Running iPhone", Platform: "ios", State: "Booted", Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityBoot: domain.CapabilityUnavailable}}},
	}
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	err := runner.Run(context.Background(), []string{"device", "boot", "--device", "ios-1"})
	if err == nil || !strings.Contains(err.Error(), "capability boot is unavailable") {
		t.Fatalf("unexpected boot error: %v", err)
	}
	if len(adapter.Operations) != 0 {
		t.Fatalf("unsupported boot was executed: %v", adapter.Operations)
	}
}
