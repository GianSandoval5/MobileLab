package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectDetectsMultiplePlatforms(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"pubspec.yaml": "name: sample",
		"package.json": `{"dependencies":{"react-native":"1.0.0"}}`,
		"android/app/src/main/AndroidManifest.xml": "<manifest />",
		"capacitor.config.ts":                      "export default {}",
	}
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	results, err := Project(root)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Flutter": true, "React Native": true, "Android": true, "Capacitor": true}
	for _, result := range results {
		delete(want, result.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing detections: %v (results: %v)", want, results)
	}
}

func TestProjectRejectsMalformedPackageJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Project(root); err == nil {
		t.Fatal("expected malformed package.json error")
	}
}
