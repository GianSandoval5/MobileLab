package doctor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/config"
)

type Level string

const (
	OK      Level = "ok"
	Warning Level = "warning"
)

type Check struct {
	Group   string
	Name    string
	Level   Level
	Message string
}

func Run(root string) []Check {
	checks := []Check{configuration(root)}
	checks = append(checks,
		androidTool("ADB", "adb", "platform-tools", "Install Android platform-tools to enable Android devices."),
		androidTool("Emulator", "emulator", "emulator", "Install Android Emulator to discover and boot AVDs."),
		androidSDK(),
	)
	if goruntime.GOOS == "darwin" {
		checks = append(checks,
			tool("iOS", "Xcode", "xcodebuild", "Install Xcode to enable iOS simulators."),
			simctl(),
		)
	} else {
		checks = append(checks, Check{Group: "iOS", Name: "simctl", Level: Warning, Message: "iOS tooling requires macOS + Xcode"})
	}
	checks = append(checks,
		tool("Frameworks", "Flutter", "flutter", "Optional; required only for Flutter project tooling."),
		nodeTool(),
	)
	return checks
}

func nodeTool() Check {
	path, err := exec.LookPath("node")
	if err != nil {
		return Check{Group: "Frameworks", Name: "Node.js", Level: Warning, Message: "Optional; Node.js 18+ is required only for React Native SDK tooling."}
	}
	output, err := exec.Command(path, "--version").Output()
	if err != nil {
		return Check{Group: "Frameworks", Name: "Node.js", Level: Warning, Message: "Unable to execute node --version"}
	}
	version := strings.TrimSpace(string(output))
	major, err := nodeMajor(version)
	if err != nil || major < 18 {
		return Check{Group: "Frameworks", Name: "Node.js", Level: Warning, Message: fmt.Sprintf("%s; Node.js 18+ is required for React Native SDK tooling", version)}
	}
	return Check{Group: "Frameworks", Name: "Node.js", Level: OK, Message: fmt.Sprintf("%s (%s)", version, path)}
}

func nodeMajor(version string) (int, error) {
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	major, _, _ := strings.Cut(value, ".")
	return strconv.Atoi(major)
}

func simctl() Check {
	path, err := exec.Command("xcrun", "-f", "simctl").Output()
	if err != nil {
		return Check{Group: "iOS", Name: "simctl", Level: Warning, Message: "xcrun could not locate simctl; install/select Xcode"}
	}
	return Check{Group: "iOS", Name: "simctl", Level: OK, Message: string(bytes.TrimSpace(path))}
}

func configuration(root string) Check {
	path := filepath.Join(root, config.DefaultFilename)
	if _, err := config.Load(path); err != nil {
		return Check{Group: "Core", Name: "Configuration", Level: Warning, Message: fmt.Sprintf("%v; run 'mobilelab init' if needed", err)}
	}
	return Check{Group: "Core", Name: "Configuration", Level: OK, Message: path}
}

func tool(group, name, binary, unavailable string) Check {
	path, err := exec.LookPath(binary)
	if err != nil {
		return Check{Group: group, Name: name, Level: Warning, Message: unavailable}
	}
	return Check{Group: group, Name: name, Level: OK, Message: path}
}

func androidTool(name, binary, directory, unavailable string) Check {
	if path, err := exec.LookPath(binary); err == nil {
		return Check{Group: "Android", Name: name, Level: OK, Message: path}
	}
	filename := binary
	if goruntime.GOOS == "windows" {
		filename += ".exe"
	}
	for _, variable := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		root := strings.TrimSpace(os.Getenv(variable))
		if root == "" {
			continue
		}
		candidate := filepath.Join(root, directory, filename)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return Check{Group: "Android", Name: name, Level: OK, Message: candidate}
		}
	}
	return Check{Group: "Android", Name: name, Level: Warning, Message: unavailable}
}

func androidSDK() Check {
	for _, name := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if value := os.Getenv(name); value != "" {
			if info, err := os.Stat(value); err == nil && info.IsDir() {
				return Check{Group: "Android", Name: "Android SDK", Level: OK, Message: value}
			}
		}
	}
	return Check{Group: "Android", Name: "Android SDK", Level: Warning, Message: "ANDROID_HOME/ANDROID_SDK_ROOT does not point to an SDK"}
}
