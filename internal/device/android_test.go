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
	if !devices[0].Emulator || devices[0].Name != "Pixel_9" || devices[0].Capabilities[domain.CapabilityLocation] != domain.CapabilityAvailable {
		t.Fatalf("unexpected emulator: %#v", devices[0])
	}
	if devices[1].Capabilities[domain.CapabilityLaunch] != domain.CapabilityUnavailable {
		t.Fatalf("offline device has launch capability: %#v", devices[1])
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

type fakeRunner struct {
	path   string
	output []byte
	args   []string
	err    error
}

func (f *fakeRunner) LookPath(string) (string, error) { return f.path, f.err }
func (f *fakeRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.args = append([]string(nil), args...)
	return f.output, f.err
}
