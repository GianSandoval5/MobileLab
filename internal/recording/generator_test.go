package recording

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/scenario"
)

func TestGenerateEncodeAndParseRecordedScenario(t *testing.T) {
	recording := domain.Recording{
		Name: "login", StartedAt: time.Now().UTC(), InitialEnvironment: domain.EnvironmentState{LatencyMS: 50},
		Events: []domain.CaptureEvent{
			{Kind: domain.CaptureEnvironment, Environment: &domain.EnvironmentMutation{Action: "auth", AuthExpired: true}},
			{Kind: domain.CaptureDeepLink, DeepLink: &domain.DeepLinkCapture{URL: "myapp://login"}},
			{Kind: domain.CaptureHTTPExchange, HTTP: &domain.RequestRecord{Method: "post", Path: "/login", Status: 401}},
		},
	}
	definition, err := GenerateScenario(recording)
	if err != nil {
		t.Fatal(err)
	}
	data, err := EncodeScenario(definition)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := (scenario.YAMLParser{}).Parse(data)
	if err != nil {
		t.Fatalf("generated YAML did not round trip:\n%s\n%v", data, err)
	}
	if len(parsed.Steps) != 3 || parsed.Steps[0].Kind != domain.StepExpireAuth || parsed.Steps[1].Value != "myapp://login" || parsed.Steps[2].Kind != domain.StepWaitForHTTP || parsed.Steps[2].Value != "POST /login 401" {
		t.Fatalf("unexpected steps: %#v", parsed.Steps)
	}
	if len(parsed.Assertions) != 2 || parsed.Assertions[1].Response.Status != 401 {
		t.Fatalf("unexpected assertions: %#v", parsed.Assertions)
	}
}

func TestWriteScenarioIsAtomicAndRequiresForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobilelab", "scenarios", "login.yaml")
	if err := WriteScenario(path, []byte("name: first\n"), false); err != nil {
		t.Fatal(err)
	}
	if err := WriteScenario(path, []byte("name: second\n"), false); err == nil {
		t.Fatal("expected overwrite rejection")
	}
	if err := WriteScenario(path, []byte("name: second\n"), true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "name: second\n" {
		t.Fatalf("unexpected file: %q %v", data, err)
	}
}
