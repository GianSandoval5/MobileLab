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

func TestRunnerWaitsForRepeatedHTTPExchangesInOrder(t *testing.T) {
	now := time.Now().UTC()
	environment := &fakeEnvironment{records: []domain.RequestRecord{
		{Method: "GET", Path: "/profile", Status: 200, Timestamp: now.Add(time.Millisecond)},
		{Method: "GET", Path: "/profile", Status: 200, Timestamp: now.Add(2 * time.Millisecond)},
	}}
	definition := domain.ScenarioDefinition{Name: "Repeated requests", Steps: []domain.ScenarioStep{
		{Kind: domain.StepWaitForHTTP, Value: "GET /profile 200"},
		{Kind: domain.StepSetLatency, Value: "25"},
		{Kind: domain.StepWaitForHTTP, Value: "GET /profile 200"},
	}}
	result, err := (Runner{Environment: environment, Device: &device.FakeAdapter{}, Now: func() time.Time { return now }}).Run(
		context.Background(), definition, domain.ScenarioRunOptions{Timeout: time.Millisecond},
	)
	if err != nil || !result.Passed || len(result.Steps) != 3 || environment.latency != 25 {
		t.Fatalf("unexpected result=%#v environment=%#v err=%v", result, environment, err)
	}
}

func TestEvaluateAssertionsCorrelatesResponseWithPreviousRequest(t *testing.T) {
	records := []domain.RequestRecord{
		{Method: "POST", Path: "/payments", Status: 200},
		{Method: "GET", Path: "/unrelated", Status: 500},
	}
	assertions := []domain.ScenarioAssertion{
		{Request: &domain.RequestExpectation{Method: "POST", Path: "/payments"}},
		{Response: &domain.ResponseExpectation{Status: 500}},
	}
	checks := evaluateAssertions(records, nil, assertions)
	if len(checks) != 2 || !checks[0].Passed || checks[1].Passed {
		t.Fatalf("unrelated HTTP 500 satisfied correlated assertion: %#v", checks)
	}
	if checks[1].Name != "POST /payments returned HTTP 500" {
		t.Fatalf("correlation is not visible in report: %#v", checks[1])
	}
}

func TestEvaluateAssertionsAcceptsCorrelatedResponse(t *testing.T) {
	records := []domain.RequestRecord{
		{Method: "POST", Path: "/payments/123", Status: 500},
		{Method: "GET", Path: "/unrelated", Status: 200},
	}
	assertions := []domain.ScenarioAssertion{
		{Request: &domain.RequestExpectation{Method: "POST", Path: "/payments/{id}"}},
		{Response: &domain.ResponseExpectation{Status: 500}},
	}
	checks := evaluateAssertions(records, nil, assertions)
	if !allPassed(checks) {
		t.Fatalf("correlated response did not pass: %#v", checks)
	}
}

func TestEvaluateAppEventAssertionsUsesOnlyMatchingFrameworkAndPassState(t *testing.T) {
	failed := false
	passed := true
	events := []domain.AppEvent{
		{Framework: domain.FrameworkReactNative, Kind: domain.AppEventAssertion, Name: "checkout.ready", Passed: &passed},
		{Framework: domain.FrameworkFlutter, Kind: domain.AppEventAssertion, Name: "checkout.ready", Passed: &failed},
	}
	assertions := []domain.ScenarioAssertion{{AppEvent: &domain.AppEventExpectation{
		Framework: domain.FrameworkFlutter, Kind: domain.AppEventAssertion, Name: "checkout.ready",
	}}}
	checks := evaluateAssertions(nil, events, assertions)
	if len(checks) != 1 || checks[0].Passed || checks[0].Message != "app assertion reported passed=false" {
		t.Fatalf("unexpected app assertion result: %#v", checks)
	}
}

func TestRunnerIgnoresAppEventsBeforeScenarioStart(t *testing.T) {
	now := time.Now().UTC()
	environment := &fakeEnvironment{appEvents: []domain.AppEvent{{
		Framework: domain.FrameworkFlutter, Kind: domain.AppEventMarker, Name: "ready", Timestamp: now.Add(-time.Second),
	}}}
	definition := domain.ScenarioDefinition{Name: "Fresh app marker", Assertions: []domain.ScenarioAssertion{{
		AppEvent: &domain.AppEventExpectation{Kind: domain.AppEventMarker, Name: "ready"},
	}}}
	result, err := (Runner{Environment: environment, Device: &device.FakeAdapter{}, Now: func() time.Time { return now }}).Run(
		context.Background(), definition, domain.ScenarioRunOptions{Timeout: time.Millisecond},
	)
	if err != nil || result.Passed || len(result.Assertions) != 1 || result.Assertions[0].Passed {
		t.Fatalf("stale app event satisfied scenario: result=%#v err=%v", result, err)
	}
}

type fakeEnvironment struct {
	latency     int
	errorStatus int
	authExpired bool
	resets      int
	records     []domain.RequestRecord
	appEvents   []domain.AppEvent
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
func (f *fakeEnvironment) RecentAppEvents(context.Context, int) ([]domain.AppEvent, error) {
	return f.appEvents, nil
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
