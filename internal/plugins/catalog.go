package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Descriptor struct {
	Manifest   Manifest `json:"manifest"`
	ProjectDir string   `json:"project_dir"`
	Directory  string   `json:"directory"`
	Executable string   `json:"executable"`
	SHA256     string   `json:"sha256"`
}

type Issue struct {
	Path string
	Err  error
}

type Catalog struct {
	ProjectDir string
}

func (catalog Catalog) Root() string {
	return filepath.Join(catalog.ProjectDir, "mobilelab", "plugins")
}

func (catalog Catalog) Discover() ([]Descriptor, []Issue, error) {
	entries, err := os.ReadDir(catalog.Root())
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read plugin directory %q: %w", catalog.Root(), err)
	}
	descriptors := make([]Descriptor, 0, len(entries))
	issues := make([]Issue, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		descriptor, err := catalog.Load(entry.Name())
		if err != nil {
			issues = append(issues, Issue{Path: filepath.Join(catalog.Root(), entry.Name()), Err: err})
			continue
		}
		descriptors = append(descriptors, descriptor)
	}
	sort.Slice(descriptors, func(i, j int) bool { return descriptors[i].Manifest.Name < descriptors[j].Manifest.Name })
	return descriptors, issues, nil
}

func (catalog Catalog) Load(name string) (Descriptor, error) {
	if !pluginNamePattern.MatchString(name) {
		return Descriptor{}, fmt.Errorf("plugin name must match %s", pluginNamePattern)
	}
	directory := filepath.Join(catalog.Root(), name)
	manifest, err := LoadManifest(filepath.Join(directory, ManifestFilename))
	if err != nil {
		return Descriptor{}, err
	}
	if manifest.Name != name {
		return Descriptor{}, fmt.Errorf("plugin directory %q must match manifest name %q", name, manifest.Name)
	}
	executable, err := resolveExecutable(directory, manifest.Executable)
	if err != nil {
		return Descriptor{}, err
	}
	fingerprint, err := fileSHA256(executable)
	if err != nil {
		return Descriptor{}, err
	}
	projectDir, err := filepath.Abs(catalog.ProjectDir)
	if err != nil {
		return Descriptor{}, fmt.Errorf("resolve project directory: %w", err)
	}
	return Descriptor{Manifest: manifest, ProjectDir: projectDir, Directory: directory, Executable: executable, SHA256: fingerprint}, nil
}

func resolveExecutable(directory, name string) (string, error) {
	candidate := filepath.Join(directory, name)
	if runtime.GOOS == "windows" && filepath.Ext(candidate) == "" {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			candidate += ".exe"
		}
	}
	realDirectory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", fmt.Errorf("resolve plugin directory %q: %w", directory, err)
	}
	realExecutable, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve plugin executable %q: %w", candidate, err)
	}
	relative, err := filepath.Rel(realDirectory, realExecutable)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("plugin executable must remain inside %q", directory)
	}
	info, err := os.Stat(realExecutable)
	if err != nil {
		return "", fmt.Errorf("inspect plugin executable %q: %w", realExecutable, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("plugin executable %q must be a regular file", realExecutable)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("plugin executable %q is not executable", realExecutable)
	}
	return realExecutable, nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open plugin executable for fingerprint: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("fingerprint plugin executable: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
