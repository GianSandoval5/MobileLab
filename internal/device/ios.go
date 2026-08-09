package device

import (
	"context"
	"encoding/json"
	"fmt"
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
	return parseSimctlDevices(output)
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

func (a *IOSAdapter) SetNetworkCondition(context.Context, string, domain.NetworkCondition) error {
	return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityNetworkOffline, Reason: "simctl does not provide portable network conditioning"}
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
			UDID        string `json:"udid"`
			Name        string `json:"name"`
			State       string `json:"state"`
			IsAvailable bool   `json:"isAvailable"`
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
			devices = append(devices, domain.Device{
				ID: candidate.UDID, Name: candidate.Name, Platform: "ios", Emulator: true, State: candidate.State,
				Details: map[string]string{"runtime": runtimeName},
				Capabilities: map[domain.Capability]domain.CapabilityLevel{
					domain.CapabilityLaunch: level, domain.CapabilityStop: level, domain.CapabilityClear: level, domain.CapabilityBoot: boot, domain.CapabilityDeepLink: level,
					domain.CapabilityLocation: level, domain.CapabilityNetworkOffline: domain.CapabilityUnavailable,
					domain.CapabilityNetworkLatency: domain.CapabilityUnavailable, domain.CapabilityPush: domain.CapabilityUnavailable,
				},
			})
		}
	}
	return devices, nil
}
