package doctor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"

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
		tool("Android", "ADB", "adb", "Install Android platform-tools to enable Android devices."),
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
		tool("Frameworks", "Node.js", "node", "Optional; the MobileLab core does not require Node.js."),
	)
	return checks
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
