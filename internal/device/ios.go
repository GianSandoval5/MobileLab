package device

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	goruntime "runtime"
	"strconv"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type IOSAdapter struct {
	runner ProcessRunner
}

func NewIOSAdapter(runner ProcessRunner) *IOSAdapter {
	return &IOSAdapter{runner: runner}
}

func (a *IOSAdapter) Platform() string { return "ios" }

func (a *IOSAdapter) Detect(ctx context.Context) ([]domain.Device, error) {
	if goruntime.GOOS != "darwin" {
		return nil, nil
	}
	xcrun, err := a.runner.LookPath("xcrun")
	if err != nil {
		return nil, nil
	}
	output, err := a.runner.Run(ctx, xcrun, "simctl", "list", "devices", "available", "--json")
	if err != nil {
		return nil, err
	}
	devices, err := parseSimctlDevices(output)
	if err != nil {
		return nil, err
	}
	if _, pushErr := a.runner.Run(ctx, xcrun, "simctl", "help", "push"); pushErr == nil {
		for index := range devices {
			if strings.EqualFold(devices[index].State, "Booted") {
				devices[index].Capabilities[domain.CapabilityPush] = domain.CapabilityAvailable
			}
		}
	}
	return devices, nil
}

func (a *IOSAdapter) LaunchApp(ctx context.Context, deviceID, appID string) error {
	return a.runSimctl(ctx, domain.CapabilityLaunch, "launch", deviceID, appID)
}

func (a *IOSAdapter) StopApp(ctx context.Context, deviceID, appID string) error {
	return a.runSimctl(ctx, domain.CapabilityStop, "terminate", deviceID, appID)
}

func (a *IOSAdapter) ClearApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("iOS bundle ID is required")
	}
	return a.runSimctl(ctx, domain.CapabilityClear, "uninstall", deviceID, appID)
}

func (a *IOSAdapter) BootDevice(ctx context.Context, deviceID string) error {
	return a.runSimctl(ctx, domain.CapabilityBoot, "boot", deviceID)
}

func (a *IOSAdapter) OpenDeepLink(ctx context.Context, deviceID, value string) error {
	return a.runSimctl(ctx, domain.CapabilityDeepLink, "openurl", deviceID, value)
}

func (a *IOSAdapter) SetLocation(ctx context.Context, deviceID string, location domain.Location) error {
	value := strconv.FormatFloat(location.Latitude, 'f', 6, 64) + "," + strconv.FormatFloat(location.Longitude, 'f', 6, 64)
	return a.runSimctl(ctx, domain.CapabilityLocation, "location", deviceID, "set", value)
}

func (a *IOSAdapter) SetNetworkCondition(_ context.Context, _ string, condition domain.NetworkCondition) error {
	return domain.CapabilityError{Platform: a.Platform(), Capability: networkCapability(condition), Reason: "simctl does not provide portable network conditioning"}
}

func (a *IOSAdapter) SendPush(ctx context.Context, deviceID, appID string, notification domain.PushNotification) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("iOS bundle ID is required")
	}
	payload, err := iosPushPayload(notification)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "mobilelab-push-*.apns")
	if err != nil {
		return fmt.Errorf("create temporary APNs payload: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temporary APNs payload: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary APNs payload: %w", err)
	}
	return a.runSimctl(ctx, domain.CapabilityPush, "push", deviceID, appID, path)
}

func iosPushPayload(notification domain.PushNotification) ([]byte, error) {
	alert := map[string]string{}
	if notification.Title != "" {
		alert["title"] = notification.Title
	}
	if notification.Body != "" {
		alert["body"] = notification.Body
	}
	payload := make(map[string]any, len(notification.Data)+1)
	for key, value := range notification.Data {
		if key == "aps" {
			return nil, fmt.Errorf("push data key %q is reserved by APNs", key)
		}
		payload[key] = value
	}
	payload["aps"] = map[string]any{"alert": alert, "sound": "default"}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode APNs payload: %w", err)
	}
	if len(encoded) > 4096 {
		return nil, fmt.Errorf("APNs payload is %d bytes; simctl supports at most 4096", len(encoded))
	}
	return encoded, nil
}

func (a *IOSAdapter) runSimctl(ctx context.Context, capability domain.Capability, args ...string) error {
	if goruntime.GOOS != "darwin" {
		return domain.CapabilityError{Platform: a.Platform(), Capability: capability, Reason: "iOS tooling requires macOS and Xcode"}
	}
	xcrun, err := a.runner.LookPath("xcrun")
	if err != nil {
		return domain.CapabilityError{Platform: a.Platform(), Capability: capability, Reason: "xcrun was not found"}
	}
	_, err = a.runner.Run(ctx, xcrun, append([]string{"simctl"}, args...)...)
	return err
}

func parseSimctlDevices(output []byte) ([]domain.Device, error) {
	var payload struct {
		Devices map[string][]struct {
			UDID                 string `json:"udid"`
			Name                 string `json:"name"`
			State                string `json:"state"`
			IsAvailable          bool   `json:"isAvailable"`
			DeviceTypeIdentifier string `json:"deviceTypeIdentifier"`
			LastBootedAt         string `json:"lastBootedAt"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parse simctl devices: %w", err)
	}
	var devices []domain.Device
	for runtimeName, candidates := range payload.Devices {
		for _, candidate := range candidates {
			if !candidate.IsAvailable {
				continue
			}
			level := domain.CapabilityUnavailable
			if strings.EqualFold(candidate.State, "Booted") {
				level = domain.CapabilityAvailable
			}
			boot := domain.CapabilityUnavailable
			if strings.EqualFold(candidate.State, "Shutdown") {
				boot = domain.CapabilityAvailable
			}
			details := map[string]string{"runtime": runtimeName}
			if version := iosRuntimeVersion(runtimeName); version != "" {
				details["osVersion"] = version
			}
			if candidate.DeviceTypeIdentifier != "" {
				details["deviceTypeIdentifier"] = candidate.DeviceTypeIdentifier
			}
			if candidate.LastBootedAt != "" {
				details["lastBootedAt"] = candidate.LastBootedAt
			}
			devices = append(devices, domain.Device{
				ID: candidate.UDID, Name: candidate.Name, Platform: "ios", Emulator: true, State: candidate.State,
				Details: details,
				Capabilities: map[domain.Capability]domain.CapabilityLevel{
					domain.CapabilityLaunch: level, domain.CapabilityStop: level, domain.CapabilityClear: level, domain.CapabilityBoot: boot, domain.CapabilityDeepLink: level,
					domain.CapabilityLocation: level, domain.CapabilityNetworkOffline: domain.CapabilityUnavailable,
					domain.CapabilityNetworkOnline:  domain.CapabilityUnavailable,
					domain.CapabilityNetworkLatency: domain.CapabilityUnavailable, domain.CapabilityPush: domain.CapabilityUnavailable,
				},
			})
		}
	}
	return devices, nil
}

func iosRuntimeVersion(runtimeName string) string {
	const marker = ".iOS-"
	index := strings.LastIndex(runtimeName, marker)
	if index < 0 {
		return ""
	}
	return strings.ReplaceAll(strings.TrimPrefix(runtimeName[index:], marker), "-", ".")
}
