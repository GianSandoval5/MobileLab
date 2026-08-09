package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/app"
	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/detect"
	"github.com/mobilelab-dev/mobilelab/internal/device"
	"github.com/mobilelab-dev/mobilelab/internal/doctor"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	openapiimport "github.com/mobilelab-dev/mobilelab/internal/openapi"
	"github.com/mobilelab-dev/mobilelab/internal/recording"
	"github.com/mobilelab-dev/mobilelab/internal/reporting"
	"github.com/mobilelab-dev/mobilelab/internal/runtime"
	"github.com/mobilelab-dev/mobilelab/internal/scenario"
)

var Version = "0.7.0-dev"

type Runner struct {
	Out            io.Writer
	Err            io.Writer
	Dir            string
	DeviceAdapters []domain.DeviceAdapter
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
	case "network":
		return r.network(ctx, args[1:])
	case "push":
		return r.push(ctx, args[1:])
	case "device":
		return r.deviceCommand(ctx, args[1:])
	case "run":
		return r.runScenario(ctx, args[1:])
	case "scenario":
		return r.scenarioCommand(ctx, args[1:])
	case "record":
		return r.record(ctx, args[1:])
	case "replay":
		if len(args) < 2 {
			return fmt.Errorf("usage: mobilelab replay <name> [scenario options]")
		}
		return r.scenarioCommand(ctx, append([]string{"run"}, args[1:]...))
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
	fmt.Fprintf(r.Out, "MobileLab is running\n✓ PID       %d\n✓ Uptime    %s\n✓ Requests  %d\n✓ App events %d\n✓ Scenarios %d\n", status.PID, status.Uptime, status.Requests, status.AppEvents, status.ScenarioRuns)
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
	toolchains, err := detect.Toolchains()
	if err != nil {
		return fmt.Errorf("unable to detect framework tooling: %w", err)
	}
	fmt.Fprintln(r.Out, "\nAvailable framework tooling:")
	if len(toolchains) == 0 {
		fmt.Fprintln(r.Out, "- None (optional; the MobileLab core remains available)")
	}
	for _, result := range toolchains {
		fmt.Fprintf(r.Out, "✓ %-14s %s\n", result.Name, strings.Join(result.Evidence, ", "))
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
		for _, capability := range deviceCapabilityOrder() {
			fmt.Fprintf(r.Out, "  %-18s %s\n", capability, detected.Capabilities[capability])
		}
	}
	return nil
}

func (r Runner) deepLink(ctx context.Context, args []string) error {
	if len(args) < 2 || args[0] != "open" {
		return fmt.Errorf("usage: mobilelab deeplink open <url> [--platform android|ios] [--device id]")
	}
	selection, _, err := parseDeviceFlags(args[2:], false, false)
	if err != nil {
		return err
	}
	adapter, detected, err := r.selectDevice(ctx, selection, domain.CapabilityDeepLink)
	if err != nil {
		return err
	}
	if err := adapter.OpenDeepLink(ctx, detected.ID, args[1]); err != nil {
		return fmt.Errorf("unable to open deep link on %s: %w", detected.Name, err)
	}
	if err := (runtime.Client{ConfigPath: r.configPath()}).RecordCapture(ctx, domain.CaptureEvent{
		Kind:     domain.CaptureDeepLink,
		DeepLink: &domain.DeepLinkCapture{URL: args[1], Platform: detected.Platform, DeviceID: detected.ID},
	}); err != nil {
		fmt.Fprintf(r.Err, "! Deep link was not recorded: %v\n", err)
	}
	fmt.Fprintf(r.Out, "Deep link opened on %s.\n", detected.Name)
	return nil
}

func (r Runner) location(ctx context.Context, args []string) error {
	if len(args) < 3 || args[0] != "set" {
		return fmt.Errorf("usage: mobilelab location set <latitude> <longitude> [--platform android|ios] [--device id]")
	}
	latitude, latErr := strconv.ParseFloat(args[1], 64)
	longitude, lonErr := strconv.ParseFloat(args[2], 64)
	if latErr != nil || lonErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return fmt.Errorf("latitude must be -90..90 and longitude must be -180..180")
	}
	selection, _, err := parseDeviceFlags(args[3:], false, false)
	if err != nil {
		return err
	}
	adapter, detected, err := r.selectDevice(ctx, selection, domain.CapabilityLocation)
	if err != nil {
		return err
	}
	if err := adapter.SetLocation(ctx, detected.ID, domain.Location{Latitude: latitude, Longitude: longitude}); err != nil {
		return fmt.Errorf("unable to set location on %s: %w", detected.Name, err)
	}
	fmt.Fprintf(r.Out, "Location set on %s.\n", detected.Name)
	return nil
}

func (r Runner) network(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab network <online|offline|slow> [--platform android|ios] [--device id]")
	}
	condition := domain.NetworkCondition(args[0])
	capability := networkConditionCapability(condition)
	if capability == "" {
		return fmt.Errorf("network condition must be online, offline, or slow")
	}
	selection, _, err := parseDeviceFlags(args[1:], false, false)
	if err != nil {
		return err
	}
	adapter, detected, err := r.selectDevice(ctx, selection, capability)
	if err != nil {
		return err
	}
	if err := adapter.SetNetworkCondition(ctx, detected.ID, condition); err != nil {
		return fmt.Errorf("unable to set network %s on %s: %w", condition, detected.Name, err)
	}
	fmt.Fprintf(r.Out, "Network set to %s on %s.\n", condition, detected.Name)
	return nil
}

func (r Runner) push(ctx context.Context, args []string) error {
	if len(args) < 2 || args[0] != "send" {
		return fmt.Errorf("usage: mobilelab push send <fixture> --app-id <bundle-id> [--platform ios] [--device id]")
	}
	selection, options, err := parseDeviceFlags(args[2:], true, false)
	if err != nil {
		return err
	}
	if options.AppID == "" {
		return fmt.Errorf("push send requires --app-id <bundle-id>")
	}
	cfg, err := config.Load(filepath.Join(r.Dir, config.DefaultFilename))
	if err != nil {
		return err
	}
	fixture, exists := cfg.Push[args[1]]
	if !exists {
		return fmt.Errorf("push fixture %q was not found in %s", args[1], config.DefaultFilename)
	}
	adapter, detected, err := r.selectDevice(ctx, selection, domain.CapabilityPush)
	if err != nil {
		return err
	}
	notification := domain.PushNotification{Title: fixture.Title, Body: fixture.Body, Data: fixture.Data}
	if err := adapter.SendPush(ctx, detected.ID, options.AppID, notification); err != nil {
		return fmt.Errorf("unable to send push %s to %s: %w", args[1], detected.Name, err)
	}
	fmt.Fprintf(r.Out, "Push %s sent to %s.\n", args[1], detected.Name)
	return nil
}

func networkConditionCapability(condition domain.NetworkCondition) domain.Capability {
	switch condition {
	case domain.NetworkOnline:
		return domain.CapabilityNetworkOnline
	case domain.NetworkOffline:
		return domain.CapabilityNetworkOffline
	case domain.NetworkSlow:
		return domain.CapabilityNetworkLatency
	default:
		return ""
	}
}

type deviceSelection struct {
	Platform string
	ID       string
}

type deviceCommandOptions struct {
	Selection deviceSelection
	AppID     string
	JSON      bool
}

func (r Runner) deviceCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab device list | info | launch | stop | clear | boot")
	}
	switch args[0] {
	case "list":
		selection, options, err := parseDeviceFlags(args[1:], false, true)
		if err != nil {
			return err
		}
		devices := r.devices(ctx)
		filtered := devices[:0]
		for _, detected := range devices {
			platformMatches := selection.Platform == "" || detected.Platform == selection.Platform
			deviceMatches := selection.ID == "" || detected.ID == selection.ID
			if platformMatches && deviceMatches {
				filtered = append(filtered, detected)
			}
		}
		if options.JSON {
			encoder := json.NewEncoder(r.Out)
			encoder.SetIndent("", "  ")
			return encoder.Encode(filtered)
		}
		if len(filtered) == 0 {
			fmt.Fprintln(r.Out, "No devices detected.")
			return nil
		}
		for _, detected := range filtered {
			fmt.Fprintf(r.Out, "%-8s %-24s %-38s %s\n", detected.Platform, detected.Name, detected.ID, detected.State)
		}
		return nil
	case "info":
		selection, options, err := parseDeviceFlags(args[1:], false, true)
		if err != nil {
			return err
		}
		_, detected, err := r.selectDevice(ctx, selection, "")
		if err != nil {
			return err
		}
		if options.JSON {
			encoder := json.NewEncoder(r.Out)
			encoder.SetIndent("", "  ")
			return encoder.Encode(detected)
		}
		fmt.Fprintf(r.Out, "Device: %s\nPlatform: %s\nID: %s\nState: %s\nEmulator: %t\n", detected.Name, detected.Platform, detected.ID, detected.State, detected.Emulator)
		if len(detected.Details) > 0 {
			keys := make([]string, 0, len(detected.Details))
			for key := range detected.Details {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			fmt.Fprintln(r.Out, "Details:")
			for _, key := range keys {
				fmt.Fprintf(r.Out, "  %-22s %s\n", key, detected.Details[key])
			}
		}
		fmt.Fprintln(r.Out, "Capabilities:")
		for _, capability := range deviceCapabilityOrder() {
			fmt.Fprintf(r.Out, "  %-22s %s\n", capability, detected.Capabilities[capability])
		}
		return nil
	case "launch", "stop", "clear":
		selection, options, err := parseDeviceFlags(args[1:], true, false)
		if err != nil {
			return err
		}
		if options.AppID == "" {
			return fmt.Errorf("device %s requires --app-id <application-id>", args[0])
		}
		capability := map[string]domain.Capability{"launch": domain.CapabilityLaunch, "stop": domain.CapabilityStop, "clear": domain.CapabilityClear}[args[0]]
		adapter, detected, err := r.selectDevice(ctx, selection, capability)
		if err != nil {
			return err
		}
		switch args[0] {
		case "launch":
			err = adapter.LaunchApp(ctx, detected.ID, options.AppID)
		case "stop":
			err = adapter.StopApp(ctx, detected.ID, options.AppID)
		case "clear":
			err = adapter.ClearApp(ctx, detected.ID, options.AppID)
		}
		if err != nil {
			return fmt.Errorf("device %s failed on %s: %w", args[0], detected.Name, err)
		}
		fmt.Fprintf(r.Out, "Device %s completed on %s.\n", args[0], detected.Name)
		return nil
	case "boot":
		selection, _, err := parseDeviceFlags(args[1:], false, false)
		if err != nil {
			return err
		}
		adapter, detected, err := r.selectDevice(ctx, selection, domain.CapabilityBoot)
		if err != nil {
			return err
		}
		if err := adapter.BootDevice(ctx, detected.ID); err != nil {
			return fmt.Errorf("boot device %s: %w", detected.Name, err)
		}
		fmt.Fprintf(r.Out, "Boot requested for %s.\n", detected.Name)
		return nil
	default:
		return fmt.Errorf("unknown device command %q; use list, info, launch, stop, clear, or boot", args[0])
	}
}

func deviceCapabilityOrder() []domain.Capability {
	return []domain.Capability{
		domain.CapabilityBoot, domain.CapabilityLaunch, domain.CapabilityStop, domain.CapabilityClear,
		domain.CapabilityDeepLink, domain.CapabilityLocation, domain.CapabilityNetworkOnline,
		domain.CapabilityNetworkOffline, domain.CapabilityNetworkLatency, domain.CapabilityPush,
	}
}

func parseDeviceFlags(args []string, allowAppID, allowJSON bool) (deviceSelection, deviceCommandOptions, error) {
	options := deviceCommandOptions{}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		if flag == "--json" && allowJSON {
			options.JSON = true
			continue
		}
		if index+1 >= len(args) {
			return deviceSelection{}, deviceCommandOptions{}, fmt.Errorf("option %q requires a value", flag)
		}
		value := args[index+1]
		index++
		switch flag {
		case "--platform":
			if value != "android" && value != "ios" {
				return deviceSelection{}, deviceCommandOptions{}, fmt.Errorf("platform must be android or ios")
			}
			options.Selection.Platform = value
		case "--device":
			options.Selection.ID = value
		case "--app-id":
			if !allowAppID {
				return deviceSelection{}, deviceCommandOptions{}, fmt.Errorf("option --app-id is not valid for this command")
			}
			options.AppID = value
		default:
			return deviceSelection{}, deviceCommandOptions{}, fmt.Errorf("unknown device option %q", flag)
		}
	}
	return options.Selection, options, nil
}

func (r Runner) runScenario(ctx context.Context, args []string) error {
	options, err := parseRunOptions(args)
	if err != nil {
		return err
	}
	inputs, err := loadScenarioInputs(options.Path)
	if err != nil {
		return err
	}
	inputInfo, err := os.Stat(options.Path)
	if err != nil {
		return fmt.Errorf("inspect scenario input %q: %w", options.Path, err)
	}
	options.Directory = inputInfo.IsDir()
	started := time.Now().UTC()
	results := make([]domain.ScenarioResult, 0, len(inputs))
	for _, input := range inputs {
		adapter, deviceID, err := r.scenarioDevice(ctx, input.Definition, options.Platform, options.DeviceID)
		if err != nil {
			results = append(results, domain.ScenarioResult{
				Name:      input.Definition.Name,
				StartedAt: time.Now().UTC(),
				Error:     fmt.Sprintf("prepare scenario %q: %v", input.Path, err),
			})
			continue
		}
		runner := scenario.Runner{
			Environment: runtime.Client{ConfigPath: r.configPath()},
			Device:      adapter,
			Runs:        runtime.Client{ConfigPath: r.configPath()},
		}
		result, _ := runner.Run(ctx, input.Definition, domain.ScenarioRunOptions{
			DeviceID: deviceID,
			AppID:    options.AppID,
			Timeout:  options.Timeout,
		})
		results = append(results, result)
		if ctx.Err() != nil {
			break
		}
	}
	suiteName := filepath.Base(filepath.Clean(options.Path))
	if len(inputs) == 1 {
		suiteName = inputs[0].Definition.Name
	}
	suite := domain.NewScenarioSuiteResult(suiteName, started, time.Since(started).Milliseconds(), results)
	if err := r.writeScenarioReport(options, suite); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !suite.Passed {
		return fmt.Errorf("scenario suite failed: %d of %d scenario(s) failed", suite.Summary.Failed, suite.Summary.Total)
	}
	return nil
}

type runOptions struct {
	Path      string
	Platform  string
	DeviceID  string
	AppID     string
	Report    reporting.Format
	Output    string
	Timeout   time.Duration
	Directory bool
}

func parseRunOptions(args []string) (runOptions, error) {
	if len(args) == 0 {
		return runOptions{}, fmt.Errorf("usage: mobilelab run <scenario.yaml|directory> [--platform android|ios|fake] [--device id] [--app-id id] [--timeout 10s] [--report terminal|json|junit|html] [--output path]")
	}
	options := runOptions{Path: args[0], Report: reporting.FormatTerminal}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if !strings.HasPrefix(arg, "--") || index+1 >= len(args) {
			return runOptions{}, fmt.Errorf("option %q requires a value", arg)
		}
		value := args[index+1]
		index++
		switch arg {
		case "--platform":
			if value != "android" && value != "ios" && value != "fake" {
				return runOptions{}, fmt.Errorf("platform must be android, ios, or fake")
			}
			options.Platform = value
		case "--device":
			options.DeviceID = value
		case "--app-id":
			options.AppID = value
		case "--report":
			options.Report = reporting.Format(strings.ToLower(value))
		case "--output":
			options.Output = value
		case "--timeout":
			timeout, err := time.ParseDuration(value)
			if err != nil || timeout <= 0 || timeout > time.Hour {
				return runOptions{}, fmt.Errorf("timeout must be greater than zero and no more than 1h")
			}
			options.Timeout = timeout
		default:
			return runOptions{}, fmt.Errorf("unknown run option %q", arg)
		}
	}
	if !options.Report.Valid() {
		return runOptions{}, fmt.Errorf("report must be terminal, json, junit, or html")
	}
	return options, nil
}

type scenarioInput struct {
	Path       string
	Definition domain.ScenarioDefinition
}

func loadScenarioInputs(path string) ([]scenarioInput, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect scenario input %q: %w", path, err)
	}
	paths := []string{path}
	if info.IsDir() {
		paths = nil
		err = filepath.WalkDir(path, func(candidate string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			extension := strings.ToLower(filepath.Ext(entry.Name()))
			if extension == ".yaml" || extension == ".yml" {
				paths = append(paths, candidate)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("discover scenarios in %q: %w", path, err)
		}
		sort.Strings(paths)
		if len(paths) == 0 {
			return nil, fmt.Errorf("scenario directory %q contains no .yaml or .yml files", path)
		}
	} else if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("scenario input %q must be a regular file or directory", path)
	}
	inputs := make([]scenarioInput, 0, len(paths))
	for _, scenarioPath := range paths {
		data, err := os.ReadFile(scenarioPath)
		if err != nil {
			return nil, fmt.Errorf("read scenario %q: %w", scenarioPath, err)
		}
		definition, err := (scenario.YAMLParser{}).Parse(data)
		if err != nil {
			return nil, fmt.Errorf("scenario %q: %w", scenarioPath, err)
		}
		inputs = append(inputs, scenarioInput{Path: scenarioPath, Definition: definition})
	}
	return inputs, nil
}

func (r Runner) writeScenarioReport(options runOptions, suite domain.ScenarioSuiteResult) error {
	reporter, err := reporting.NewSuiteReporter(options.Report)
	if err != nil {
		return err
	}
	write := func(writer io.Writer) error {
		if options.Report == reporting.FormatJSON && len(suite.Scenarios) == 1 && !options.Directory {
			return reporting.WriteScenarioJSON(writer, suite.Scenarios[0])
		}
		return reporter.Write(writer, suite)
	}
	if options.Output == "" {
		if err := write(r.Out); err != nil {
			return fmt.Errorf("write %s report: %w", options.Report, err)
		}
		return nil
	}
	parent := filepath.Dir(options.Output)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create report directory %q: %w", parent, err)
	}
	output, err := os.OpenFile(options.Output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create report %q: %w", options.Output, err)
	}
	if err := write(output); err != nil {
		_ = output.Close()
		return fmt.Errorf("write %s report: %w", options.Report, err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close report %q: %w", options.Output, err)
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

func (r Runner) record(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab record <name> [--duration 30s] [--force]")
	}
	name := strings.TrimSuffix(args[0], ".yaml")
	if filepath.Base(name) != name || name == "." || name == ".." {
		return fmt.Errorf("recording name must be a safe file name")
	}
	force := false
	var duration time.Duration
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--force":
			force = true
		case "--duration":
			if index+1 >= len(args) {
				return fmt.Errorf("--duration requires a value such as 30s")
			}
			index++
			parsed, err := time.ParseDuration(args[index])
			if err != nil || parsed <= 0 || parsed > 24*time.Hour {
				return fmt.Errorf("duration must be greater than zero and no more than 24h")
			}
			duration = parsed
		default:
			return fmt.Errorf("unknown record option %q", args[index])
		}
	}
	path := filepath.Join(r.Dir, "mobilelab", "scenarios", name+".yaml")
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("scenario %s already exists; use --force to replace it", filepath.Base(path))
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	client := runtime.Client{ConfigPath: r.configPath()}
	started, err := client.StartRecording(ctx, name)
	if err != nil {
		return fmt.Errorf("start recording: %w", err)
	}
	fmt.Fprintf(r.Out, "Recording %s started at %s.\n", started.Name, started.StartedAt.Local().Format(time.RFC3339))
	if duration == 0 {
		fmt.Fprintln(r.Out, "Exercise the application now; press Ctrl+C to stop and generate the scenario.")
		<-ctx.Done()
	} else {
		fmt.Fprintf(r.Out, "Capturing for %s...\n", duration)
		timer := time.NewTimer(duration)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
	stopContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	captured, err := client.StopRecording(stopContext)
	if err != nil {
		return fmt.Errorf("stop recording: %w", err)
	}
	if len(captured.Events) == 0 {
		return fmt.Errorf("recording captured no events; no scenario was written")
	}
	definition, err := recording.GenerateScenario(captured)
	if err != nil {
		return err
	}
	data, err := recording.EncodeScenario(definition)
	if err != nil {
		return err
	}
	if err := recording.WriteScenario(path, data, force); err != nil {
		return err
	}
	fmt.Fprintf(r.Out, "Recorded %d event(s) to %s. Replay with 'mobilelab replay %s'.\n", len(captured.Events), path, name)
	return nil
}

func (r Runner) scenarioDevice(ctx context.Context, definition domain.ScenarioDefinition, platform, requestedID string) (domain.DeviceAdapter, string, error) {
	needsDevice := definition.Device.Network != ""
	for _, step := range definition.Steps {
		if step.Kind == domain.StepLaunchApp || step.Kind == domain.StepStopApp || step.Kind == domain.StepOpenDeepLink {
			needsDevice = true
			break
		}
	}
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
	if r.DeviceAdapters != nil {
		return r.DeviceAdapters
	}
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

func (r Runner) selectDevice(ctx context.Context, selection deviceSelection, capability domain.Capability) (domain.DeviceAdapter, domain.Device, error) {
	var matchingDevice *domain.Device
	for _, adapter := range r.adapters() {
		if selection.Platform != "" && adapter.Platform() != selection.Platform {
			continue
		}
		devices, err := adapter.Detect(ctx)
		if err != nil {
			return nil, domain.Device{}, fmt.Errorf("detect %s devices: %w", adapter.Platform(), err)
		}
		for _, detected := range devices {
			if selection.ID != "" && detected.ID != selection.ID {
				continue
			}
			copy := detected
			matchingDevice = &copy
			level := detected.Capabilities[capability]
			if capability == "" || level == domain.CapabilityAvailable || level == domain.CapabilityPartial {
				return adapter, detected, nil
			}
		}
	}
	if matchingDevice != nil {
		return nil, domain.Device{}, fmt.Errorf("%s is detected but capability %s is unavailable", matchingDevice.Name, capability)
	}
	description := "matching"
	if selection.ID != "" {
		description = selection.ID
	} else if selection.Platform != "" {
		description = selection.Platform
	}
	return nil, domain.Device{}, fmt.Errorf("no %s device found; run 'mobilelab device list'", description)
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
  network    Shape supported emulator network characteristics
  push       Send a configured local push where supported
  device     List/select devices and control app/device lifecycle
  run        Execute a scenario file/directory with terminal, JSON, JUnit, or HTML reports
  scenario   List, run, and inspect persistent scenario history
  record     Capture app traffic/actions into a portable YAML scenario
  replay     Execute a recorded scenario through the standard runner
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
