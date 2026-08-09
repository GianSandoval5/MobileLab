package scenario

import (
	"context"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/device"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestRunnerConfiguresEnvironmentExecutesPortableStepsAndAsserts(t *testing.T) {
	now := time.Now().UTC()
	environment := &fakeEnvironment{records: []domain.RequestRecord{{
		Method: "POST", Path: "/payments", Status: 500, Timestamp: now.Add(time.Millisecond),
	}}}
	runs := &fakeRunRepository{}
	adapter := &device.FakeAdapter{}
	definition := domain.ScenarioDefinition{
		Name: "Payment failure", Backend: domain.ScenarioBackend{LatencyMS: 50, Error: 500}, Auth: domain.ScenarioAuth{Token: "expired"},
		Steps: []domain.ScenarioStep{{Kind: domain.StepLaunchApp}, {Kind: domain.StepOpenDeepLink, Value: "myapp://payments"}},
		Assertions: []domain.ScenarioAssertion{
			{Request: &domain.RequestExpectation{Method: "POST", Path: "/payments"}},
			{Response: &domain.ResponseExpectation{Status: 500}},
		},
	}
	runner := Runner{Environment: environment, Device: adapter, Runs: runs, Now: func() time.Time { return now }}
	result, err := runner.Run(context.Background(), definition, domain.ScenarioRunOptions{DeviceID: "fake-1", AppID: "dev.mobilelab.app", Timeout: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed {
		t.Fatalf("scenario failed: %#v", result)
	}
	if environment.latency != 50 || environment.errorStatus != 500 || !environment.authExpired || environment.resets != 2 {
		t.Fatalf("unexpected environment calls: %#v", environment)
	}
	if len(adapter.Operations) != 2 || adapter.Operations[0] != "launch:fake-1:dev.mobilelab.app" {
		t.Fatalf("unexpected device operations: %v", adapter.Operations)
	}
	if len(runs.results) != 1 || !runs.results[0].Passed {
		t.Fatalf("scenario result was not persisted: %#v", runs.results)
	}
}

func TestRunnerReportsMissingApplicationID(t *testing.T) {
	environment := &fakeEnvironment{}
	definition := domain.ScenarioDefinition{Name: "Launch", Steps: []domain.ScenarioStep{{Kind: domain.StepLaunchApp}}}
	result, err := (Runner{Environment: environment, Device: &device.FakeAdapter{}}).Run(context.Background(), definition, domain.ScenarioRunOptions{})
	if err == nil || result.Passed || len(result.Steps) != 1 {
		t.Fatalf("expected reported launch failure, got result=%#v error=%v", result, err)
	}
}

type fakeEnvironment struct {
	latency     int
	errorStatus int
	authExpired bool
	resets      int
	records     []domain.RequestRecord
}

func (f *fakeEnvironment) SetLatency(_ context.Context, value int) error {
	f.latency = value
	return nil
}
func (f *fakeEnvironment) SetError(_ context.Context, value int) error {
	f.errorStatus = value
	return nil
}
func (f *fakeEnvironment) SetAuthExpired(_ context.Context, value bool) error {
	f.authExpired = value
	return nil
}
func (f *fakeEnvironment) Reset(context.Context) error { f.resets++; return nil }
func (f *fakeEnvironment) RecentRequests(context.Context, int) ([]domain.RequestRecord, error) {
	return f.records, nil
}

type fakeRunRepository struct {
	results []domain.ScenarioResult
	err     error
}

func (f *fakeRunRepository) Save(_ context.Context, result domain.ScenarioResult) error {
	f.results = append(f.results, result)
	return f.err
}

func (f *fakeRunRepository) Recent(context.Context, int) ([]domain.ScenarioResult, error) {
	return append([]domain.ScenarioResult(nil), f.results...), f.err
}
