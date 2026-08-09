package detect

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Result struct {
	Name     string
	Evidence []string
}

type Detector interface {
	Detect(root string) (Result, bool, error)
}

func Toolchains() ([]Result, error) {
	var results []Result
	if path, err := exec.LookPath("flutter"); err == nil {
		results = append(results, Result{Name: "Flutter SDK", Evidence: []string{path}})
	}
	if path, err := exec.LookPath("node"); err == nil {
		version, versionErr := exec.Command(path, "--version").Output()
		evidence := path
		if versionErr == nil {
			evidence = strings.TrimSpace(string(version)) + " (" + path + ")"
		}
		results = append(results, Result{Name: "Node.js", Evidence: []string{evidence}})
	}
	return results, nil
}

type fileDetector struct {
	name  string
	paths []string
}

func (d fileDetector) Detect(root string) (Result, bool, error) {
	var evidence []string
	for _, path := range d.paths {
		if _, err := os.Stat(filepath.Join(root, path)); err == nil {
			evidence = append(evidence, path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return Result{}, false, fmt.Errorf("inspect %s: %w", path, err)
		}
	}
	return Result{Name: d.name, Evidence: evidence}, len(evidence) > 0, nil
}

type reactNativeDetector struct{}

func (reactNativeDetector) Detect(root string) (Result, bool, error) {
	path := filepath.Join(root, "package.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, fmt.Errorf("read package.json: %w", err)
	}
	var manifest struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Result{}, false, fmt.Errorf("parse package.json: %w", err)
	}
	_, dependency := manifest.Dependencies["react-native"]
	_, devDependency := manifest.DevDependencies["react-native"]
	return Result{Name: "React Native", Evidence: []string{"package.json:react-native"}}, dependency || devDependency, nil
}

type appleProjectDetector struct{}

func (appleProjectDetector) Detect(root string) (Result, bool, error) {
	patterns := []string{"*.xcodeproj", "*.xcworkspace", "ios/*.xcodeproj", "ios/*.xcworkspace"}
	var evidence []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return Result{}, false, err
		}
		for _, match := range matches {
			relative, _ := filepath.Rel(root, match)
			evidence = append(evidence, relative)
		}
	}
	return Result{Name: "iOS", Evidence: evidence}, len(evidence) > 0, nil
}

func Project(root string) ([]Result, error) {
	detectors := []Detector{
		fileDetector{name: "Flutter", paths: []string{"pubspec.yaml"}},
		reactNativeDetector{},
		fileDetector{name: "Android", paths: []string{"build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts", "android/build.gradle", "android/build.gradle.kts", "android/settings.gradle", "android/settings.gradle.kts", "AndroidManifest.xml", "android/app/src/main/AndroidManifest.xml"}},
		appleProjectDetector{},
		fileDetector{name: "Ionic", paths: []string{"ionic.config.json"}},
		fileDetector{name: "Capacitor", paths: []string{"capacitor.config.ts", "capacitor.config.json", "capacitor.config.js"}},
	}

	var results []Result
	for _, detector := range detectors {
		result, found, err := detector.Detect(root)
		if err != nil {
			return nil, err
		}
		if found {
			sort.Strings(result.Evidence)
			results = append(results, result)
		}
	}
	return results, nil
}
