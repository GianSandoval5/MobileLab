package device

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type AndroidAdapter struct {
	runner ProcessRunner
}

func NewAndroidAdapter(runner ProcessRunner) *AndroidAdapter {
	return &AndroidAdapter{runner: runner}
}

func (a *AndroidAdapter) Platform() string { return "android" }

func (a *AndroidAdapter) Detect(ctx context.Context) ([]domain.Device, error) {
	adb, err := a.runner.LookPath("adb")
	if err != nil {
		return nil, nil
	}
	output, err := a.runner.Run(ctx, adb, "devices", "-l")
	if err != nil {
		return nil, err
	}
	return parseADBDevices(string(output)), nil
}

func (a *AndroidAdapter) LaunchApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("Android application ID is required")
	}
	return a.runADB(ctx, deviceID, "shell", "monkey", "-p", appID, "-c", "android.intent.category.LAUNCHER", "1")
}

func (a *AndroidAdapter) StopApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("Android application ID is required")
	}
	return a.runADB(ctx, deviceID, "shell", "am", "force-stop", appID)
}

func (a *AndroidAdapter) OpenDeepLink(ctx context.Context, deviceID, value string) error {
	if _, err := url.ParseRequestURI(value); err != nil {
		return fmt.Errorf("invalid deep link %q: %w", value, err)
	}
	return a.runADB(ctx, deviceID, "shell", "am", "start", "-W", "-a", "android.intent.action.VIEW", "-d", value)
}

func (a *AndroidAdapter) SetLocation(ctx context.Context, deviceID string, location domain.Location) error {
	if !strings.HasPrefix(deviceID, "emulator-") {
		return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityLocation, Reason: "location injection is limited to Android emulators"}
	}
	longitude := strconv.FormatFloat(location.Longitude, 'f', 6, 64)
	latitude := strconv.FormatFloat(location.Latitude, 'f', 6, 64)
	return a.runADB(ctx, deviceID, "emu", "geo", "fix", longitude, latitude)
}

func (a *AndroidAdapter) SetNetworkCondition(context.Context, string, domain.NetworkCondition) error {
	return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityNetworkOffline, Reason: "reliable device-wide network control is not implemented"}
}

func (a *AndroidAdapter) runADB(ctx context.Context, deviceID string, args ...string) error {
	adb, err := a.runner.LookPath("adb")
	if err != nil {
		return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityLaunch, Reason: "adb was not found"}
	}
	if deviceID != "" {
		args = append([]string{"-s", deviceID}, args...)
	}
	_, err = a.runner.Run(ctx, adb, args...)
	return err
}

func parseADBDevices(output string) []domain.Device {
	var devices []domain.Device
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "*") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		details := make(map[string]string)
		for _, field := range fields[2:] {
			parts := strings.SplitN(field, ":", 2)
			if len(parts) == 2 {
				details[parts[0]] = parts[1]
			}
		}
		name := details["model"]
		if name == "" {
			name = fields[0]
		}
		emulator := strings.HasPrefix(fields[0], "emulator-")
		location := domain.CapabilityUnavailable
		if emulator && fields[1] == "device" {
			location = domain.CapabilityAvailable
		}
		level := domain.CapabilityUnavailable
		if fields[1] == "device" {
			level = domain.CapabilityAvailable
		}
		devices = append(devices, domain.Device{
			ID: fields[0], Name: name, Platform: "android", Emulator: emulator, State: fields[1], Details: details,
			Capabilities: map[domain.Capability]domain.CapabilityLevel{
				domain.CapabilityLaunch: level, domain.CapabilityStop: level, domain.CapabilityDeepLink: level,
				domain.CapabilityLocation: location, domain.CapabilityNetworkOffline: domain.CapabilityUnavailable,
				domain.CapabilityNetworkLatency: domain.CapabilityUnavailable, domain.CapabilityPush: domain.CapabilityUnavailable,
			},
		})
	}
	return devices
}
