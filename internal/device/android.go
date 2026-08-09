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
	var devices []domain.Device
	var adb string
	if resolvedADB, err := a.runner.LookPath("adb"); err == nil {
		adb = resolvedADB
		output, runErr := a.runner.Run(ctx, adb, "devices", "-l")
		if runErr != nil {
			return nil, runErr
		}
		devices = parseADBDevices(string(output))
	}

	emulator, err := a.runner.LookPath("emulator")
	if err != nil {
		return devices, nil
	}
	output, err := a.runner.Run(ctx, emulator, "-list-avds")
	if err != nil {
		return nil, err
	}
	running := make(map[string]struct{})
	if adb != "" {
		for _, detected := range devices {
			if !detected.Emulator || detected.State != "device" {
				continue
			}
			name, queryErr := a.runner.Run(ctx, adb, "-s", detected.ID, "emu", "avd", "name")
			if queryErr == nil {
				if avd := parseRunningAVDName(string(name)); avd != "" {
					running[avd] = struct{}{}
				}
			}
		}
	}
	return append(devices, parseAndroidAVDs(string(output), running)...), nil
}

func (a *AndroidAdapter) LaunchApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("Android application ID is required")
	}
	return a.runADB(ctx, domain.CapabilityLaunch, deviceID, "shell", "monkey", "-p", appID, "-c", "android.intent.category.LAUNCHER", "1")
}

func (a *AndroidAdapter) StopApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("Android application ID is required")
	}
	return a.runADB(ctx, domain.CapabilityStop, deviceID, "shell", "am", "force-stop", appID)
}

func (a *AndroidAdapter) ClearApp(ctx context.Context, deviceID, appID string) error {
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("Android application ID is required")
	}
	return a.runADB(ctx, domain.CapabilityClear, deviceID, "shell", "pm", "clear", appID)
}

func (a *AndroidAdapter) BootDevice(ctx context.Context, deviceID string) error {
	const prefix = "avd:"
	if !strings.HasPrefix(deviceID, prefix) || strings.TrimSpace(strings.TrimPrefix(deviceID, prefix)) == "" {
		return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityBoot, Reason: "select a configured AVD reported by 'mobilelab device list'"}
	}
	emulator, err := a.runner.LookPath("emulator")
	if err != nil {
		return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityBoot, Reason: "Android emulator was not found"}
	}
	return a.runner.Start(ctx, emulator, "-avd", strings.TrimPrefix(deviceID, prefix))
}

func (a *AndroidAdapter) OpenDeepLink(ctx context.Context, deviceID, value string) error {
	if _, err := url.ParseRequestURI(value); err != nil {
		return fmt.Errorf("invalid deep link %q: %w", value, err)
	}
	return a.runADB(ctx, domain.CapabilityDeepLink, deviceID, "shell", "am", "start", "-W", "-a", "android.intent.action.VIEW", "-d", value)
}

func (a *AndroidAdapter) SetLocation(ctx context.Context, deviceID string, location domain.Location) error {
	if !strings.HasPrefix(deviceID, "emulator-") {
		return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityLocation, Reason: "location injection is limited to Android emulators"}
	}
	longitude := strconv.FormatFloat(location.Longitude, 'f', 6, 64)
	latitude := strconv.FormatFloat(location.Latitude, 'f', 6, 64)
	return a.runADB(ctx, domain.CapabilityLocation, deviceID, "emu", "geo", "fix", longitude, latitude)
}

func (a *AndroidAdapter) SetNetworkCondition(context.Context, string, domain.NetworkCondition) error {
	return domain.CapabilityError{Platform: a.Platform(), Capability: domain.CapabilityNetworkOffline, Reason: "reliable device-wide network control is not implemented"}
}

func (a *AndroidAdapter) runADB(ctx context.Context, capability domain.Capability, deviceID string, args ...string) error {
	adb, err := a.runner.LookPath("adb")
	if err != nil {
		return domain.CapabilityError{Platform: a.Platform(), Capability: capability, Reason: "adb was not found"}
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
				domain.CapabilityLaunch: level, domain.CapabilityStop: level, domain.CapabilityClear: level, domain.CapabilityBoot: domain.CapabilityUnavailable, domain.CapabilityDeepLink: level,
				domain.CapabilityLocation: location, domain.CapabilityNetworkOffline: domain.CapabilityUnavailable,
				domain.CapabilityNetworkLatency: domain.CapabilityUnavailable, domain.CapabilityPush: domain.CapabilityUnavailable,
			},
		})
	}
	return devices
}

func parseAndroidAVDs(output string, running map[string]struct{}) []domain.Device {
	var devices []domain.Device
	seen := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if _, active := running[name]; active {
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		seen[name] = struct{}{}
		devices = append(devices, domain.Device{
			ID:       "avd:" + name,
			Name:     name,
			Platform: "android",
			Emulator: true,
			State:    "shutdown",
			Details:  map[string]string{"avd": name},
			Capabilities: map[domain.Capability]domain.CapabilityLevel{
				domain.CapabilityLaunch: domain.CapabilityUnavailable, domain.CapabilityStop: domain.CapabilityUnavailable,
				domain.CapabilityClear: domain.CapabilityUnavailable, domain.CapabilityBoot: domain.CapabilityAvailable,
				domain.CapabilityDeepLink: domain.CapabilityUnavailable, domain.CapabilityLocation: domain.CapabilityUnavailable,
				domain.CapabilityNetworkOffline: domain.CapabilityUnavailable, domain.CapabilityNetworkLatency: domain.CapabilityUnavailable,
				domain.CapabilityPush: domain.CapabilityUnavailable,
			},
		})
	}
	return devices
}

func parseRunningAVDName(output string) string {
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(line)
		if name != "" && name != "OK" && !strings.HasPrefix(name, "KO:") {
			return name
		}
	}
	return ""
}
