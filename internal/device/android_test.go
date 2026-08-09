package device

import (
	"context"
	"reflect"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestParseADBDevicesAndCapabilities(t *testing.T) {
	output := `List of devices attached
emulator-5554 device product:sdk_gphone model:Pixel_9 device:emu64
R58M offline usb:1-1 model:Physical_Phone
`
	devices := parseADBDevices(output)
	if len(devices) != 2 {
		t.Fatalf("got %d devices: %#v", len(devices), devices)
	}
	if !devices[0].Emulator || devices[0].Name != "Pixel_9" || devices[0].Capabilities[domain.CapabilityLocation] != domain.CapabilityPartial {
		t.Fatalf("unexpected emulator: %#v", devices[0])
	}
	if devices[1].Capabilities[domain.CapabilityLaunch] != domain.CapabilityUnavailable {
		t.Fatalf("offline device has launch capability: %#v", devices[1])
	}
}

func TestAndroidPropertiesEnrichDeviceInformation(t *testing.T) {
	properties := parseAndroidProperties(`[ro.product.manufacturer]: [Google]
[ro.product.model]: [Pixel 10]
[ro.build.version.release]: [16]
[ro.build.version.sdk]: [36]
[ro.product.cpu.abi]: [arm64-v8a]
[unrelated.secret]: [ignored]
`)
	device := domain.Device{Name: "generic", Details: map[string]string{"transport_id": "1"}}
	enrichAndroidDetails(&device, properties)
	if device.Name != "Pixel 10" || device.Details["apiLevel"] != "36" || device.Details["abi"] != "arm64-v8a" {
		t.Fatalf("unexpected enriched device: %#v", device)
	}
	if _, leaked := device.Details["unrelated.secret"]; leaked {
		t.Fatalf("unexpected property leaked: %#v", device.Details)
	}
}

func TestAndroidDeepLinkUsesVerifiedADBShape(t *testing.T) {
	runner := &fakeRunner{path: "/sdk/adb"}
	adapter := NewAndroidAdapter(runner)
	if err := adapter.OpenDeepLink(context.Background(), "emulator-5554", "myapp://payments/123"); err != nil {
		t.Fatal(err)
	}
	want := []string{"-s", "emulator-5554", "shell", "am", "start", "-W", "-a", "android.intent.action.VIEW", "-d", "myapp://payments/123"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("args = %v, want %v", runner.args, want)
	}
}

func TestAndroidLocationRejectsPhysicalDevice(t *testing.T) {
	adapter := NewAndroidAdapter(&fakeRunner{path: "/sdk/adb"})
	err := adapter.SetLocation(context.Background(), "physical", domain.Location{Latitude: -5.1, Longitude: -80.6})
	if err == nil {
		t.Fatal("expected unsupported location error")
	}
}

func TestAndroidClearUsesPackageManagerForExplicitApp(t *testing.T) {
	runner := &fakeRunner{path: "/sdk/adb"}
	adapter := NewAndroidAdapter(runner)
	if err := adapter.ClearApp(context.Background(), "emulator-5554", "dev.mobilelab.app"); err != nil {
		t.Fatal(err)
	}
	want := []string{"-s", "emulator-5554", "shell", "pm", "clear", "dev.mobilelab.app"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("args = %v, want %v", runner.args, want)
	}
}

func TestParseAndroidAVDsExposesBootTargets(t *testing.T) {
	devices := parseAndroidAVDs("Pixel_9_API_35\nTablet_API_34\nPixel_9_API_35\n", nil)
	if len(devices) != 2 {
		t.Fatalf("got %d AVDs: %#v", len(devices), devices)
	}
	if devices[0].ID != "avd:Pixel_9_API_35" || devices[0].State != "shutdown" {
		t.Fatalf("unexpected AVD: %#v", devices[0])
	}
	if devices[0].Capabilities[domain.CapabilityBoot] != domain.CapabilityAvailable || devices[0].Capabilities[domain.CapabilityLaunch] != domain.CapabilityUnavailable {
		t.Fatalf("unexpected AVD capabilities: %#v", devices[0].Capabilities)
	}
}

func TestParseAndroidAVDsOmitsAlreadyRunningAVD(t *testing.T) {
	devices := parseAndroidAVDs("Pixel_9_API_35\nTablet_API_34\n", map[string]struct{}{"Pixel_9_API_35": {}})
	if len(devices) != 1 || devices[0].Name != "Tablet_API_34" {
		t.Fatalf("unexpected inactive AVDs: %#v", devices)
	}
	if got := parseRunningAVDName("Pixel_9_API_35\nOK\n"); got != "Pixel_9_API_35" {
		t.Fatalf("running AVD name = %q", got)
	}
}

func TestAndroidBootStartsExplicitAVDWithoutBlocking(t *testing.T) {
	runner := &fakeRunner{path: "/sdk/emulator"}
	adapter := NewAndroidAdapter(runner)
	if err := adapter.BootDevice(context.Background(), "avd:Pixel_9_API_35"); err != nil {
		t.Fatal(err)
	}
	want := []string{"-avd", "Pixel_9_API_35"}
	if !runner.started || !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("started=%t args=%v, want %v", runner.started, runner.args, want)
	}
}

func TestAndroidBootRejectsAttachedDeviceID(t *testing.T) {
	adapter := NewAndroidAdapter(&fakeRunner{path: "/sdk/emulator"})
	if err := adapter.BootDevice(context.Background(), "emulator-5554"); err == nil {
		t.Fatal("expected attached device ID to be rejected")
	}
}

func TestAndroidNetworkSlowUsesOfficialEmulatorConsoleCommands(t *testing.T) {
	runner := &fakeRunner{path: "/sdk/adb"}
	adapter := NewAndroidAdapter(runner)
	if err := adapter.SetNetworkCondition(context.Background(), "emulator-5554", domain.NetworkSlow); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"-s", "emulator-5554", "emu", "network", "delay", "gprs"},
		{"-s", "emulator-5554", "emu", "network", "speed", "gprs"},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %v, want %v", runner.calls, want)
	}
}

func TestAndroidNetworkOfflineRemainsUnavailable(t *testing.T) {
	adapter := NewAndroidAdapter(&fakeRunner{path: "/sdk/adb"})
	if err := adapter.SetNetworkCondition(context.Background(), "emulator-5554", domain.NetworkOffline); err == nil {
		t.Fatal("expected offline to remain unavailable")
	}
}

type fakeRunner struct {
	path    string
	output  []byte
	args    []string
	calls   [][]string
	err     error
	started bool
}

func (f *fakeRunner) Start(_ context.Context, _ string, args ...string) error {
	f.args = append([]string(nil), args...)
	f.started = true
	return f.err
}

func (f *fakeRunner) LookPath(string) (string, error) { return f.path, f.err }
func (f *fakeRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.args = append([]string(nil), args...)
	f.calls = append(f.calls, append([]string(nil), args...))
	return f.output, f.err
}
