package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/mobilelab-dev/mobilelab/internal/app"
	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/detect"
	"github.com/mobilelab-dev/mobilelab/internal/device"
	"github.com/mobilelab-dev/mobilelab/internal/doctor"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	openapiimport "github.com/mobilelab-dev/mobilelab/internal/openapi"
	"github.com/mobilelab-dev/mobilelab/internal/reporting"
	"github.com/mobilelab-dev/mobilelab/internal/runtime"
	"github.com/mobilelab-dev/mobilelab/internal/scenario"
)

var Version = "0.1.0-dev"

type Runner struct {
	Out io.Writer
	Err io.Writer
	Dir string
}

func New(out, errOut io.Writer, dir string) Runner {
	return Runner{Out: out, Err: errOut, Dir: dir}
}

func (r Runner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		r.help()
		return nil
	}
	switch args[0] {
	case "help", "--help", "-h":
		r.help()
		return nil
	case "version", "--version", "-v":
		fmt.Fprintf(r.Out, "MobileLab %s\n", Version)
		return nil
	case "init":
		return r.init(ctx, args[1:])
	case "detect":
		return r.detect(ctx)
	case "doctor":
		return r.doctor()
	case "capabilities":
		return r.capabilities(ctx)
	case "deeplink":
		return r.deepLink(ctx, args[1:])
	case "location":
		return r.location(ctx, args[1:])
	case "run":
		return r.runScenario(ctx, args[1:])
	case "scenario":
		return r.scenarioCommand(ctx, args[1:])
	case "start":
		return r.start(ctx, args[1:])
	case "status":
		return r.status(ctx)
	case "stop":
		return r.stop(ctx)
	case "api":
		return r.api(ctx, args[1:])
	case "auth":
		return r.auth(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\nRun 'mobilelab help' to see available commands", args[0])
	}
}

func (r Runner) start(ctx context.Context, args []string) error {
	headless := false
	for _, arg := range args {
		if arg == "--headless" {
			headless = true
			continue
		}
		return fmt.Errorf("unknown start option %q", arg)
	}
	environment, err := runtime.NewEnvironment(r.configPath(), r.Out)
	if err != nil {
		return fmt.Errorf("unable to start MobileLab: %w", err)
	}
	environment.SetHeadless(headless)
	return environment.Run(ctx)
}

func (r Runner) status(ctx context.Context) error {
	status, err := runtime.GetStatus(ctx, r.configPath())
	if err != nil {
		return fmt.Errorf("unable to get MobileLab status: %w", err)
	}
	fmt.Fprintf(r.Out, "MobileLab is running\n✓ PID       %d\n✓ Uptime    %s\n✓ Requests  %d\n✓ Scenarios %d\n", status.PID, status.Uptime, status.Requests, status.ScenarioRuns)
	fmt.Fprintf(r.Out, "  Latency   %dms\n", status.LatencyMS)
	if status.Error != 0 {
		fmt.Fprintf(r.Out, "! Error     HTTP %d forced\n", status.Error)
	} else {
		fmt.Fprintln(r.Out, "✓ Errors    inactive")
	}
	if status.AuthExpired {
		fmt.Fprintln(r.Out, "! Auth      expired")
	} else {
		fmt.Fprintln(r.Out, "✓ Auth      active")
	}
	return nil
}

func (r Runner) stop(ctx context.Context) error {
	if err := runtime.Stop(ctx, r.configPath()); err != nil {
		return fmt.Errorf("unable to stop MobileLab: %w", err)
	}
	fmt.Fprintln(r.Out, "MobileLab shutdown requested.")
	return nil
}

func (r Runner) api(ctx context.Context, args []string) error {
	if len(args) == 2 && args[0] == "import" {
		result, err := openapiimport.Import(ctx, args[1], r.configPath())
		if err != nil {
			return fmt.Errorf("unable to import OpenAPI: %w", err)
		}
		fmt.Fprintf(r.Out, "%d endpoints detected\n%d schemas detected\n%d initial scenarios generated\n", result.Endpoints, result.Schemas, result.ScenariosGenerated)
		return nil
	}
	if len(args) == 1 && args[0] == "reset" {
		if err := runtime.ResetFaults(ctx, r.configPath()); err != nil {
			return err
		}
		fmt.Fprintln(r.Out, "API faults reset.")
		return nil
	}
	if len(args) != 2 || (args[0] != "latency" && args[0] != "error") {
		return fmt.Errorf("usage: mobilelab api import <openapi.yaml> | latency <milliseconds> | error <status> | reset")
	}
	value, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("%s value must be a number", args[0])
	}
	if args[0] == "latency" {
		if value < 0 || value > 300_000 {
			return fmt.Errorf("latency must be between 0 and 300000 milliseconds")
		}
		err = runtime.SetLatency(ctx, r.configPath(), value)
		if err == nil {
			fmt.Fprintf(r.Out, "Global API latency set to %dms.\n", value)
		}
		return err
	}
	if value < 400 || value > 599 {
		return fmt.Errorf("error status must be between 400 and 599")
	}
	err = runtime.SetError(ctx, r.configPath(), value)
	if err == nil {
		fmt.Fprintf(r.Out, "API now forces HTTP %d. Run 'mobilelab api reset' to restore configured responses.\n", value)
	}
	return err
}

func (r Runner) auth(ctx context.Context, args []string) error {
	if len(args) != 1 || (args[0] != "expire" && args[0] != "reset") {
		return fmt.Errorf("usage: mobilelab auth expire | reset")
	}
	expired := args[0] == "expire"
	if err := runtime.SetAuthExpired(ctx, r.configPath(), expired); err != nil {
		return err
	}
	if expired {
		fmt.Fprintln(r.Out, "Auth session forced to expired.")
	} else {
		fmt.Fprintln(r.Out, "Auth session restored.")
	}
	return nil
}

func (r Runner) configPath() string {
	return filepath.Join(r.Dir, config.DefaultFilename)
}

func (r Runner) init(ctx context.Context, args []string) error {
	openAPIPath := ""
	if len(args) > 0 {
		if len(args) != 2 || args[0] != "--openapi" {
			return fmt.Errorf("usage: mobilelab init [--openapi path]")
		}
		openAPIPath = args[1]
	}
	fmt.Fprintln(r.Out, "Analyzing project...")
	result, err := app.Initialize(r.Dir)
	if err != nil {
		return fmt.Errorf("unable to initialize MobileLab: %w", err)
	}
	if len(result.Detected) == 0 {
		fmt.Fprintln(r.Out, "! No mobile framework detected; creating a framework-neutral environment")
	} else {
		for _, item := range result.Detected {
			fmt.Fprintf(r.Out, "✓ %s (%s)\n", item.Name, strings.Join(item.Evidence, ", "))
		}
	}
	fmt.Fprintln(r.Out, "\nGenerating MobileLab environment...")
	for _, created := range result.Created {
		fmt.Fprintf(r.Out, "✓ %s\n", created)
	}
	if openAPIPath != "" {
		imported, err := openapiimport.Import(ctx, openAPIPath, result.ConfigPath)
		if err != nil {
			return fmt.Errorf("environment created, but OpenAPI import failed: %w", err)
		}
		fmt.Fprintf(r.Out, "✓ OpenAPI: %d endpoints, %d schemas, %d scenarios generated\n", imported.Endpoints, imported.Schemas, imported.ScenariosGenerated)
	}
	fmt.Fprintln(r.Out, "\nReady. Run 'mobilelab start'.")
	return nil
}

func (r Runner) detect(ctx context.Context) error {
	results, err := detect.Project(r.Dir)
	if err != nil {
		return fmt.Errorf("unable to detect project: %w", err)
	}
	fmt.Fprintln(r.Out, "Detected project capabilities:")
	if len(results) == 0 {
		fmt.Fprintln(r.Out, "- No supported mobile project files found")
	} else {
		for _, result := range results {
			fmt.Fprintf(r.Out, "✓ %-14s %s\n", result.Name, strings.Join(result.Evidence, ", "))
		}
	}
	fmt.Fprintln(r.Out, "\nAvailable devices:")
	devices := r.devices(ctx)
	if len(devices) == 0 {
		fmt.Fprintln(r.Out, "- None (start an emulator/simulator or connect a device)")
	}
	for _, detected := range devices {
		fmt.Fprintf(r.Out, "✓ %-8s %-24s %s (%s)\n", detected.Platform, detected.Name, detected.ID, detected.State)
	}
	return nil
}

func (r Runner) doctor() error {
	checks := doctor.Run(r.Dir)
	fmt.Fprintln(r.Out, "MobileLab Doctor")
	warnings := 0
	group := ""
	for _, check := range checks {
		if check.Group != group {
			group = check.Group
			fmt.Fprintf(r.Out, "\n%s\n", group)
		}
		mark := "✓"
		if check.Level == doctor.Warning {
			mark = "!"
			warnings++
		}
		fmt.Fprintf(r.Out, "%s %-25s %s\n", mark, check.Name, check.Message)
	}
	fmt.Fprintf(r.Out, "\nDoctor completed with %d warning(s).\n", warnings)
	return nil
}

func (r Runner) capabilities(ctx context.Context) error {
	devices := r.devices(ctx)
	if len(devices) == 0 {
		return fmt.Errorf("no available device found; run 'mobilelab detect' after starting an emulator or simulator")
	}
	for _, detected := range devices {
		fmt.Fprintf(r.Out, "%s (%s)\n", detected.Name, detected.Platform)
		for _, capability := range []domain.Capability{
			domain.CapabilityLaunch, domain.CapabilityStop, domain.CapabilityDeepLink, domain.CapabilityLocation,
			domain.CapabilityNetworkOffline, domain.CapabilityNetworkLatency, domain.CapabilityPush,
		} {
			fmt.Fprintf(r.Out, "  %-18s %s\n", capability, detected.Capabilities[capability])
		}
	}
	return nil
}

func (r Runner) deepLink(ctx context.Context, args []string) error {
	if len(args) != 2 || args[0] != "open" {
		return fmt.Errorf("usage: mobilelab deeplink open <url>")
	}
	adapter, detected, err := r.firstDevice(ctx)
	if err != nil {
		return err
	}
	if err := adapter.OpenDeepLink(ctx, detected.ID, args[1]); err != nil {
		return fmt.Errorf("unable to open deep link on %s: %w", detected.Name, err)
	}
	fmt.Fprintf(r.Out, "Deep link opened on %s.\n", detected.Name)
	return nil
}

func (r Runner) location(ctx context.Context, args []string) error {
	if len(args) != 3 || args[0] != "set" {
		return fmt.Errorf("usage: mobilelab location set <latitude> <longitude>")
	}
	latitude, latErr := strconv.ParseFloat(args[1], 64)
	longitude, lonErr := strconv.ParseFloat(args[2], 64)
	if latErr != nil || lonErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return fmt.Errorf("latitude must be -90..90 and longitude must be -180..180")
	}
	adapter, detected, err := r.firstDevice(ctx)
	if err != nil {
		return err
	}
	if err := adapter.SetLocation(ctx, detected.ID, domain.Location{Latitude: latitude, Longitude: longitude}); err != nil {
		return fmt.Errorf("unable to set location on %s: %w", detected.Name, err)
	}
	fmt.Fprintf(r.Out, "Location set on %s.\n", detected.Name)
	return nil
}

func (r Runner) runScenario(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab run <scenario.yaml> [--platform android|ios|fake] [--device id] [--app-id id] [--report terminal|json] [--output path]")
	}
	path := args[0]
	options := map[string]string{"report": "terminal"}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if !strings.HasPrefix(arg, "--") || index+1 >= len(args) {
			return fmt.Errorf("option %q requires a value", arg)
		}
		key := strings.TrimPrefix(arg, "--")
		switch key {
		case "platform", "device", "app-id", "report", "output":
		default:
			return fmt.Errorf("unknown run option %q", arg)
		}
		index++
		options[key] = args[index]
	}
	if options["report"] != "terminal" && options["report"] != "json" {
		return fmt.Errorf("report must be terminal or json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read scenario %q: %w", path, err)
	}
	definition, err := (scenario.YAMLParser{}).Parse(data)
	if err != nil {
		return err
	}
	adapter, deviceID, err := r.scenarioDevice(ctx, definition, options["platform"], options["device"])
	if err != nil {
		return err
	}
	runner := scenario.Runner{
		Environment: runtime.Client{ConfigPath: r.configPath()},
		Device:      adapter,
		Runs:        runtime.Client{ConfigPath: r.configPath()},
	}
	result, runErr := runner.Run(ctx, definition, domain.ScenarioRunOptions{DeviceID: deviceID, AppID: options["app-id"]})

	writer := r.Out
	var outputFile *os.File
	if outputPath := options["output"]; outputPath != "" {
		outputFile, err = os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return fmt.Errorf("create report %q: %w", outputPath, err)
		}
		defer outputFile.Close()
		writer = outputFile
	}
	if options["report"] == "json" {
		if err := reporting.WriteScenarioJSON(writer, result); err != nil {
			return fmt.Errorf("write JSON report: %w", err)
		}
	} else {
		reporting.WriteScenarioTerminal(writer, result)
	}
	if runErr != nil {
		return runErr
	}
	if !result.Passed {
		return fmt.Errorf("scenario assertions failed")
	}
	return nil
}

func (r Runner) scenarioCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab scenario list | run <name> [options] | history [--json]")
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("usage: mobilelab scenario list")
		}
		return r.listScenarios()
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("usage: mobilelab scenario run <name> [options]")
		}
		name := args[1]
		if filepath.Ext(name) == "" {
			name += ".yaml"
		}
		path := filepath.Join(r.Dir, "mobilelab", "scenarios", filepath.Base(name))
		return r.runScenario(ctx, append([]string{path}, args[2:]...))
	case "history":
		jsonOutput := len(args) == 2 && args[1] == "--json"
		if len(args) > 2 || (len(args) == 2 && !jsonOutput) {
			return fmt.Errorf("usage: mobilelab scenario history [--json]")
		}
		results, err := (runtime.Client{ConfigPath: r.configPath()}).Recent(ctx, 50)
		if err != nil {
			return fmt.Errorf("unable to read scenario history: %w", err)
		}
		if jsonOutput {
			encoder := json.NewEncoder(r.Out)
			encoder.SetIndent("", "  ")
			return encoder.Encode(results)
		}
		if len(results) == 0 {
			fmt.Fprintln(r.Out, "No scenario runs recorded.")
			return nil
		}
		for _, result := range results {
			status := "PASS"
			if !result.Passed {
				status = "FAIL"
			}
			fmt.Fprintf(r.Out, "%s  %-40s %dms  %s\n", result.StartedAt.Local().Format("2006-01-02 15:04:05"), result.Name, result.DurationMS, status)
		}
		return nil
	default:
		return fmt.Errorf("unknown scenario command %q; use list, run, or history", args[0])
	}
}

func (r Runner) listScenarios() error {
	directory := filepath.Join(r.Dir, "mobilelab", "scenarios")
	paths, err := filepath.Glob(filepath.Join(directory, "*.yaml"))
	if err != nil {
		return fmt.Errorf("list scenarios: %w", err)
	}
	if len(paths) == 0 {
		fmt.Fprintf(r.Out, "No scenarios found in %s.\n", directory)
		return nil
	}
	fmt.Fprintln(r.Out, "Available scenarios:")
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read scenario %q: %w", path, err)
		}
		definition, err := (scenario.YAMLParser{}).Parse(data)
		if err != nil {
			return fmt.Errorf("invalid scenario %q: %w", path, err)
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		fmt.Fprintf(r.Out, "- %-28s %s\n", name, definition.Name)
	}
	return nil
}

func (r Runner) scenarioDevice(ctx context.Context, definition domain.ScenarioDefinition, platform, requestedID string) (domain.DeviceAdapter, string, error) {
	needsDevice := definition.Device.Network != "" || len(definition.Steps) > 0
	if platform == "fake" || (!needsDevice && platform == "") {
		return &device.FakeAdapter{}, requestedID, nil
	}
	for _, adapter := range r.adapters() {
		if platform != "" && adapter.Platform() != platform {
			continue
		}
		devices, err := adapter.Detect(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("detect %s devices: %w", adapter.Platform(), err)
		}
		for _, detected := range devices {
			if requestedID != "" && detected.ID != requestedID {
				continue
			}
			if detected.State == "device" || strings.EqualFold(detected.State, "Booted") {
				return adapter, detected.ID, nil
			}
		}
	}
	if platform == "" {
		platform = "Android/iOS"
	}
	return nil, "", fmt.Errorf("no ready %s device found; run 'mobilelab detect'", platform)
}

func (r Runner) adapters() []domain.DeviceAdapter {
	runner := device.ExecRunner{}
	return []domain.DeviceAdapter{device.NewAndroidAdapter(runner), device.NewIOSAdapter(runner)}
}

func (r Runner) devices(ctx context.Context) []domain.Device {
	var result []domain.Device
	for _, adapter := range r.adapters() {
		detected, err := adapter.Detect(ctx)
		if err != nil {
			fmt.Fprintf(r.Err, "! %s detection failed: %v\n", adapter.Platform(), err)
			continue
		}
		result = append(result, detected...)
	}
	return result
}

func (r Runner) firstDevice(ctx context.Context) (domain.DeviceAdapter, domain.Device, error) {
	for _, adapter := range r.adapters() {
		devices, err := adapter.Detect(ctx)
		if err != nil {
			continue
		}
		for _, detected := range devices {
			if detected.State == "device" || strings.EqualFold(detected.State, "Booted") {
				return adapter, detected, nil
			}
		}
	}
	return nil, domain.Device{}, fmt.Errorf("no ready Android device or iOS simulator found; run 'mobilelab detect'")
}

func (r Runner) help() {
	fmt.Fprintln(r.Out, `MobileLab - local mobile development and scenario testing

Usage:
  mobilelab <command>

Available commands:
  init       Detect the current project and create a MobileLab environment
  start      Start the API sandbox (use --headless to disable the dashboard)
  status     Show live environment and fault status
  stop       Gracefully stop the running environment
  api        Set/reset global API latency and errors
  auth       Expire/reset the local auth session
  detect     Detect mobile frameworks and project platforms
  doctor     Diagnose local configuration and mobile tooling
  capabilities Show only capabilities actually available on detected devices
  deeplink   Open a deep link on a ready device
  location   Set emulator/simulator location when supported
  run        Execute a portable YAML scenario and report PASS/FAIL
  scenario   List, run, and inspect persistent scenario history
  version    Print the MobileLab version
  help       Show this help`)
}

func Main(ctx context.Context, args []string) int {
	ctx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to determine the current directory: %v\n", err)
		return 1
	}
	runner := New(os.Stdout, os.Stderr, dir)
	if err := runner.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
