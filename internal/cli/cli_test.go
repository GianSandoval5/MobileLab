package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/device"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/reporting"
	protocol "github.com/mobilelab-dev/mobilelab/pkg/plugin"
	"github.com/mobilelab-dev/mobilelab/schemas"
)

type pluginProcessFunc func(context.Context, string, string, []string, []byte, int) ([]byte, error)

func (function pluginProcessFunc) Run(ctx context.Context, executable, directory string, environment []string, input []byte, limit int) ([]byte, error) {
	return function(ctx, executable, directory, environment, input, limit)
}

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

func TestSchemaCommandPrintsEmbeddedStableContracts(t *testing.T) {
	for _, kind := range []schemas.Kind{schemas.Config, schemas.Scenario} {
		var output bytes.Buffer
		runner := New(&output, &bytes.Buffer{}, t.TempDir())
		if err := runner.Run(context.Background(), []string{"schema", string(kind)}); err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(output.Bytes(), &document); err != nil {
			t.Fatalf("%s schema output is not JSON: %v", kind, err)
		}
		if document["$schema"] == nil || document["$id"] == nil {
			t.Fatalf("%s schema output is incomplete: %#v", kind, document)
		}
	}
}

func TestMigrateCommandChecksThenUpgradesLegacyProject(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default("legacy")
	configPath := filepath.Join(root, config.DefaultFilename)
	if err := config.Write(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(strings.TrimPrefix(string(data), "schema_version: 1\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	scenarioDirectory := filepath.Join(root, "mobilelab", "scenarios")
	if err := os.MkdirAll(scenarioDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	scenarioPath := filepath.Join(scenarioDirectory, "legacy.yaml")
	if err := os.WriteFile(scenarioPath, []byte("name: Legacy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	runner := New(&output, &bytes.Buffer{}, root)
	if err := runner.Run(context.Background(), []string{"migrate", "--check"}); err == nil || !strings.Contains(err.Error(), "2 document(s)") {
		t.Fatalf("expected migration check failure, got %v", err)
	}
	if err := runner.Run(context.Background(), []string{"migrate"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Migrated 2 document(s)") {
		t.Fatalf("unexpected migration output: %q", output.String())
	}
	for _, path := range []string{configPath, scenarioPath} {
		migrated, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(migrated), "schema_version: 1\n") {
			t.Fatalf("%s was not migrated: %s", path, migrated)
		}
	}
}

func TestEndpointCommandResolvesHostAndSelectedEmulator(t *testing.T) {
	root := t.TempDir()
	if err := config.Write(filepath.Join(root, config.DefaultFilename), config.Default("endpoint")); err != nil {
		t.Fatal(err)
	}
	adapter := &device.FakeAdapter{
		PlatformName: "android",
		Devices: []domain.Device{{
			ID: "emulator-5554", Name: "Pixel", Platform: "android", Emulator: true, State: "device",
		}},
	}
	var output, errors bytes.Buffer
	runner := New(&output, &errors, root)
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	if err := runner.Run(context.Background(), []string{"endpoint"}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "http://127.0.0.1:4566\n" {
		t.Fatalf("unexpected host endpoint: %q", output.String())
	}
	output.Reset()
	if err := runner.Run(context.Background(), []string{"endpoint", "--platform", "android", "--device", "emulator-5554", "--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"url": "http://10.0.2.2:4566"`) || !strings.Contains(output.String(), `"device_id": "emulator-5554"`) {
		t.Fatalf("unexpected Android endpoint: %q", output.String())
	}
}

func TestPluginCommandsDiscoverInspectAndRunExplicitPlugin(t *testing.T) {
	root := t.TempDir()
	createTestPlugin(t, root)
	var output, errors bytes.Buffer
	runner := New(&output, &errors, root)
	runner.PluginProcess = pluginProcessFunc(func(_ context.Context, executable, directory string, environment []string, input []byte, limit int) ([]byte, error) {
		realDirectory, err := filepath.EvalSymlinks(directory)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(realDirectory, executable)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("executable escaped plugin directory: %q", executable)
		}
		if limit != protocol.MaxMessageBytes {
			t.Fatalf("unexpected output limit: %d", limit)
		}
		for _, variable := range environment {
			if strings.HasPrefix(variable, "MOBILELAB_TEST_SECRET=") {
				t.Fatalf("secret environment variable was inherited: %q", variable)
			}
		}
		request, err := protocol.DecodeRequest(bytes.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if request.Action != "echo" || string(request.Input) != `{"hello":"plugins"}` || request.Context.ProjectDir != root {
			t.Fatalf("unexpected request: %#v", request)
		}
		var encoded bytes.Buffer
		if err := protocol.EncodeResponse(&encoded, protocol.Response{
			Protocol: protocol.ProtocolVersion, RequestID: request.RequestID, Success: true,
			Message: "echo complete", Output: json.RawMessage(`{"received":true}`),
		}); err != nil {
			t.Fatal(err)
		}
		return encoded.Bytes(), nil
	})

	if err := runner.Run(context.Background(), []string{"plugin", "list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "echo") || !strings.Contains(output.String(), "actions: echo") {
		t.Fatalf("unexpected plugin list: %q", output.String())
	}
	output.Reset()
	if err := runner.Run(context.Background(), []string{"plugin", "inspect", "echo", "--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"sha256"`) || !strings.Contains(output.String(), `"mobilelab.plugin/v1"`) {
		t.Fatalf("unexpected plugin inspection: %q", output.String())
	}

	inputPath := filepath.Join(root, "input.json")
	if err := os.WriteFile(inputPath, []byte(`{"hello":"plugins"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(root, "artifacts", "plugin.json")
	if err := os.Setenv("MOBILELAB_TEST_SECRET", "must-not-leak"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("MOBILELAB_TEST_SECRET") })
	if err := runner.Run(context.Background(), []string{"plugin", "run", "echo", "echo", "--input", inputPath, "--output", outputPath}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{\n  \"received\": true\n}\n" || !strings.Contains(errors.String(), "echo complete") {
		t.Fatalf("unexpected plugin result: output=%q stderr=%q", data, errors.String())
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(outputPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("plugin output permissions = %o, want 600", info.Mode().Perm())
		}
	}
}

func createTestPlugin(t *testing.T, root string) {
	t.Helper()
	directory := filepath.Join(root, "mobilelab", "plugins", "echo")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `api_version: mobilelab.plugin/v1
name: echo
version: 0.1.0
description: Echo test plugin
executable: mobilelab-plugin-echo
actions:
  - name: echo
    description: Echo structured input
`
	if err := os.WriteFile(filepath.Join(directory, "plugin.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(directory, "mobilelab-plugin-echo")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	if err := os.WriteFile(executable, []byte("test plugin"), 0o755); err != nil {
		t.Fatal(err)
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

func TestLoadScenarioInputsRecursivelySortsYAMLAndPreflightsAllFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"z.yml":               "name: Last\n",
		"nested/a.yaml":       "name: First\n",
		"nested/ignored.json": `{}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inputs, err := loadScenarioInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 2 || inputs[0].Definition.Name != "First" || inputs[1].Definition.Name != "Last" {
		t.Fatalf("unexpected inputs: %#v", inputs)
	}
	if err := os.WriteFile(filepath.Join(root, "broken.yaml"), []byte("unknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadScenarioInputs(root); err == nil || !strings.Contains(err.Error(), "broken.yaml") {
		t.Fatalf("expected preflight parse error, got %v", err)
	}
}

func TestParseRunOptionsSupportsCIReportsAndTimeout(t *testing.T) {
	options, err := parseRunOptions([]string{"scenarios", "--platform", "fake", "--timeout", "15s", "--report", "junit", "--output", "artifacts/results.xml"})
	if err != nil {
		t.Fatal(err)
	}
	if options.Report != reporting.FormatJUnit || options.Timeout != 15*time.Second || options.Platform != "fake" {
		t.Fatalf("unexpected options: %#v", options)
	}
	if _, err := parseRunOptions([]string{"scenarios", "--report", "tap"}); err == nil {
		t.Fatal("expected unsupported report rejection")
	}
}

func TestWriteScenarioReportCreatesParentDirectories(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "artifacts", "nested", "results.xml")
	suite := domain.NewScenarioSuiteResult("CI", time.Now().UTC(), 10, []domain.ScenarioResult{{
		Name: "Smoke", Passed: true, StartedAt: time.Now().UTC(), Steps: []domain.ScenarioCheck{{Name: "ready", Passed: true}},
	}})
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := runner.writeScenarioReport(runOptions{Report: reporting.FormatJUnit, Output: path}, suite); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<testsuites name="CI" tests="1" failures="0"`) {
		t.Fatalf("unexpected report: %s", data)
	}
}

func TestWriteDirectoryJSONKeepsSuiteEnvelopeForOneScenario(t *testing.T) {
	var output bytes.Buffer
	suite := domain.NewScenarioSuiteResult("scenarios", time.Now().UTC(), 10, []domain.ScenarioResult{{
		Name: "Only scenario", Passed: true, StartedAt: time.Now().UTC(),
	}})
	runner := New(&output, &bytes.Buffer{}, t.TempDir())
	if err := runner.writeScenarioReport(runOptions{Report: reporting.FormatJSON, Directory: true}, suite); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"summary"`) || !strings.Contains(output.String(), `"scenarios"`) {
		t.Fatalf("directory JSON lost suite envelope: %s", output.String())
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

func TestDeviceInfoPrintsSortedDetailsAndCapabilities(t *testing.T) {
	adapter := &device.FakeAdapter{
		PlatformName: "android",
		Devices: []domain.Device{{
			ID: "emulator-5554", Name: "Pixel", Platform: "android", State: "device", Emulator: true,
			Details:      map[string]string{"osVersion": "16", "abi": "arm64-v8a"},
			Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityLaunch: domain.CapabilityAvailable},
		}},
	}
	var output bytes.Buffer
	runner := New(&output, &output, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	if err := runner.Run(context.Background(), []string{"device", "info", "--device", "emulator-5554"}); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	if !strings.Contains(got, "Details:") || !strings.Contains(got, "abi") || !strings.Contains(got, "osVersion") || !strings.Contains(got, "Capabilities:") {
		t.Fatalf("incomplete device info: %q", got)
	}
	if strings.Index(got, "abi") > strings.Index(got, "osVersion") {
		t.Fatalf("details are not sorted: %q", got)
	}
}

func TestNetworkSlowExecutesPartialCapability(t *testing.T) {
	adapter := &device.FakeAdapter{
		PlatformName: "android",
		Devices: []domain.Device{{
			ID: "emulator-5554", Name: "Pixel", Platform: "android", State: "device",
			Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityNetworkLatency: domain.CapabilityPartial},
		}},
	}
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	if err := runner.Run(context.Background(), []string{"network", "slow", "--device", "emulator-5554"}); err != nil {
		t.Fatal(err)
	}
	if len(adapter.Operations) != 1 || adapter.Operations[0] != "network:emulator-5554:slow" {
		t.Fatalf("unexpected operations: %v", adapter.Operations)
	}
}

func TestNetworkOfflineRejectsUnadvertisedCapability(t *testing.T) {
	adapter := &device.FakeAdapter{
		PlatformName: "android",
		Devices: []domain.Device{{
			ID: "emulator-5554", Name: "Pixel", Platform: "android", State: "device",
			Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityNetworkOffline: domain.CapabilityUnavailable},
		}},
	}
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	if err := runner.Run(context.Background(), []string{"network", "offline", "--device", "emulator-5554"}); err == nil {
		t.Fatal("expected unavailable offline capability")
	}
	if len(adapter.Operations) != 0 {
		t.Fatalf("unsupported network operation ran: %v", adapter.Operations)
	}
}

func TestPushSendLoadsFixtureAndUsesSelectedSimulator(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default("sample")
	if err := config.Write(filepath.Join(root, config.DefaultFilename), cfg); err != nil {
		t.Fatal(err)
	}
	adapter := &device.FakeAdapter{
		PlatformName: "ios",
		Devices: []domain.Device{{
			ID: "ios-1", Name: "iPhone", Platform: "ios", State: "Booted",
			Capabilities: map[domain.Capability]domain.CapabilityLevel{domain.CapabilityPush: domain.CapabilityAvailable},
		}},
	}
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, root)
	runner.DeviceAdapters = []domain.DeviceAdapter{adapter}
	if err := runner.Run(context.Background(), []string{"push", "send", "payment-success", "--platform", "ios", "--device", "ios-1", "--app-id", "dev.mobilelab.app"}); err != nil {
		t.Fatal(err)
	}
	if len(adapter.Operations) != 1 || adapter.Operations[0] != "push:ios-1:dev.mobilelab.app" {
		t.Fatalf("unexpected push operations: %v", adapter.Operations)
	}
}
