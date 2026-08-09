package device

import (
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestParseSimctlDevices(t *testing.T) {
	payload := []byte(`{"devices":{"com.apple.CoreSimulator.SimRuntime.iOS-18-0":[{"udid":"A-1","name":"iPhone 16","state":"Booted","isAvailable":true},{"udid":"A-2","name":"Old","state":"Shutdown","isAvailable":false}]}}`)
	devices, err := parseSimctlDevices(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].ID != "A-1" {
		t.Fatalf("unexpected devices: %#v", devices)
	}
	if devices[0].Capabilities[domain.CapabilityDeepLink] != domain.CapabilityAvailable {
		t.Fatalf("booted simulator lacks deep link capability: %#v", devices[0])
	}
}

func TestParseSimctlShutdownDeviceExposesOnlyBootLifecycle(t *testing.T) {
	payload := []byte(`{"devices":{"runtime":[{"udid":"A-2","name":"iPhone 16","state":"Shutdown","isAvailable":true}]}}`)
	devices, err := parseSimctlDevices(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].Capabilities[domain.CapabilityBoot] != domain.CapabilityAvailable {
		t.Fatalf("shutdown simulator lacks boot capability: %#v", devices)
	}
	if devices[0].Capabilities[domain.CapabilityLaunch] != domain.CapabilityUnavailable {
		t.Fatalf("shutdown simulator incorrectly reports launch: %#v", devices[0])
	}
}
