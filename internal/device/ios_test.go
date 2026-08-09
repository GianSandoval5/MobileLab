package device

import (
	"encoding/json"
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

func TestIOSPushPayloadUsesAPNsShapeAndRejectsReservedData(t *testing.T) {
	encoded, err := iosPushPayload(domain.PushNotification{
		Title: "Payment completed",
		Body:  "Done",
		Data:  map[string]any{"transactionId": "ABC123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["transactionId"] != "ABC123" || payload["aps"] == nil {
		t.Fatalf("unexpected APNs payload: %#v", payload)
	}
	if _, err := iosPushPayload(domain.PushNotification{Body: "Bad", Data: map[string]any{"aps": "override"}}); err == nil {
		t.Fatal("expected reserved aps key to be rejected")
	}
	if _, err := iosPushPayload(domain.PushNotification{Body: "Too large", Data: map[string]any{"value": string(make([]byte, 4097))}}); err == nil {
		t.Fatal("expected oversized payload to be rejected")
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

func TestParseSimctlIncludesRuntimeDeviceInformation(t *testing.T) {
	payload := []byte(`{"devices":{"com.apple.CoreSimulator.SimRuntime.iOS-18-2":[{"udid":"A-1","name":"iPhone 16","state":"Booted","isAvailable":true,"deviceTypeIdentifier":"com.apple.CoreSimulator.SimDeviceType.iPhone-16","lastBootedAt":"2026-08-09T10:00:00Z"}]}}`)
	devices, err := parseSimctlDevices(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].Details["osVersion"] != "18.2" || devices[0].Details["deviceTypeIdentifier"] == "" {
		t.Fatalf("unexpected device details: %#v", devices)
	}
}
