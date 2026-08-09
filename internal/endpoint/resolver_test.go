package endpoint

import (
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestResolveHostAndSimulatorEndpoints(t *testing.T) {
	cfg := config.Default("endpoint")
	tests := []struct {
		platform string
		device   *domain.Device
		want     string
	}{
		{want: "http://127.0.0.1:4566"},
		{platform: "android", device: &domain.Device{ID: "emulator-5554", Platform: "android", Emulator: true}, want: "http://10.0.2.2:4566"},
		{platform: "ios", device: &domain.Device{ID: "ios-1", Platform: "ios", Emulator: true}, want: "http://127.0.0.1:4566"},
	}
	for _, test := range tests {
		result, err := Resolve(cfg, test.platform, test.device)
		if err != nil || result.URL != test.want {
			t.Fatalf("Resolve(%q) = %#v, %v; want %s", test.platform, result, err, test.want)
		}
	}
}

func TestResolvePhysicalDeviceRequiresReachableHost(t *testing.T) {
	cfg := config.Default("endpoint")
	physical := &domain.Device{ID: "physical-1", Platform: "android"}
	if _, err := Resolve(cfg, "android", physical); err == nil || !strings.Contains(err.Error(), "adb reverse") {
		t.Fatalf("expected actionable physical-device error, got %v", err)
	}
	cfg.Server.Host = "192.168.1.10"
	result, err := Resolve(cfg, "android", physical)
	if err != nil || result.URL != "http://192.168.1.10:4566" {
		t.Fatalf("unexpected trusted-network result=%#v err=%v", result, err)
	}
}

func TestResolveRejectsUnspecifiedAddressForDevices(t *testing.T) {
	cfg := config.Default("endpoint")
	cfg.Server.Host = "0.0.0.0"
	device := &domain.Device{ID: "emulator-5554", Platform: "android", Emulator: true}
	if _, err := Resolve(cfg, "android", device); err == nil || !strings.Contains(err.Error(), "bind address") {
		t.Fatalf("expected unspecified-host error, got %v", err)
	}
}
