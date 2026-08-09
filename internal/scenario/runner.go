package scenario

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type Runner struct {
	Environment domain.ScenarioEnvironment
	Device      domain.DeviceAdapter
	Runs        domain.ScenarioRunRepository
	Now         func() time.Time
}

func (r Runner) Run(ctx context.Context, definition domain.ScenarioDefinition, options domain.ScenarioRunOptions) (result domain.ScenarioResult, resultErr error) {
	if err := definition.Validate(); err != nil {
		return domain.ScenarioResult{}, err
	}
	if r.Environment == nil || r.Device == nil {
		return domain.ScenarioResult{}, fmt.Errorf("scenario environment and device adapter are required")
	}
	if r.Now == nil {
		r.Now = time.Now
	}
	if options.Timeout <= 0 {
		options.Timeout = 3 * time.Second
	}
	started := r.Now().UTC()
	result = domain.ScenarioResult{Name: definition.Name, StartedAt: started}
	defer func() {
		if err := r.Environment.Reset(context.WithoutCancel(ctx)); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("reset scenario environment: %w", err))
		}
		result.DurationMS = r.Now().Sub(started).Milliseconds()
		result.Passed = resultErr == nil && allPassed(result.Steps) && allPassed(result.Assertions)
		if resultErr != nil {
			result.Error = resultErr.Error()
		}
		if r.Runs != nil {
			if err := r.Runs.Save(context.WithoutCancel(ctx), result); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("persist scenario result: %w", err))
				result.Passed = false
				result.Error = resultErr.Error()
			}
		}
	}()

	if err := r.Environment.Reset(ctx); err != nil {
		return result, fmt.Errorf("prepare scenario environment: %w", err)
	}
	if err := r.Environment.SetLatency(ctx, definition.Backend.LatencyMS); err != nil {
		return result, fmt.Errorf("configure scenario latency: %w", err)
	}
	if definition.Backend.Error != 0 {
		if err := r.Environment.SetError(ctx, definition.Backend.Error); err != nil {
			return result, fmt.Errorf("configure scenario error: %w", err)
		}
	}
	if err := r.Environment.SetAuthExpired(ctx, definition.Auth.Token == "expired"); err != nil {
		return result, fmt.Errorf("configure scenario auth: %w", err)
	}
	if definition.Device.Network != "" {
		if err := r.Device.SetNetworkCondition(ctx, options.DeviceID, definition.Device.Network); err != nil {
			result.Steps = append(result.Steps, failedCheck("set network "+string(definition.Device.Network), err))
			return result, err
		}
		result.Steps = append(result.Steps, passedCheck("set network "+string(definition.Device.Network)))
	}

	for _, step := range definition.Steps {
		name := string(step.Kind)
		var err error
		switch step.Kind {
		case domain.StepLaunchApp:
			if options.AppID == "" {
				err = fmt.Errorf("launch_app requires an application ID")
			} else {
				err = r.Device.LaunchApp(ctx, options.DeviceID, options.AppID)
			}
		case domain.StepStopApp:
			if options.AppID == "" {
				err = fmt.Errorf("stop_app requires an application ID")
			} else {
				err = r.Device.StopApp(ctx, options.DeviceID, options.AppID)
			}
		case domain.StepOpenDeepLink:
			err = r.Device.OpenDeepLink(ctx, options.DeviceID, step.Value)
		default:
			err = fmt.Errorf("unsupported scenario step %q", step.Kind)
		}
		if err != nil {
			result.Steps = append(result.Steps, failedCheck(name, err))
			return result, fmt.Errorf("scenario step %s: %w", name, err)
		}
		result.Steps = append(result.Steps, passedCheck(name))
	}

	if len(definition.Assertions) > 0 {
		result.Assertions = r.waitForAssertions(ctx, started, definition.Assertions, options.Timeout)
	}
	return result, nil
}

func (r Runner) waitForAssertions(ctx context.Context, started time.Time, assertions []domain.ScenarioAssertion, timeout time.Duration) []domain.ScenarioCheck {
	deadline := time.Now().Add(timeout)
	for {
		records, err := r.Environment.RecentRequests(ctx, 500)
		if err != nil {
			return []domain.ScenarioCheck{failedCheck("read observed requests", err)}
		}
		var appEvents []domain.AppEvent
		if hasAppEventAssertions(assertions) {
			appEvents, err = r.Environment.RecentAppEvents(ctx, 500)
			if err != nil {
				return []domain.ScenarioCheck{failedCheck("read observed app events", err)}
			}
		}
		checks := evaluateAssertions(recordsSince(records, started), appEventsSince(appEvents, started), assertions)
		if allPassed(checks) || time.Now().After(deadline) {
			return checks
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return []domain.ScenarioCheck{failedCheck("wait for assertions", ctx.Err())}
		case <-timer.C:
		}
	}
}

func evaluateAssertions(records []domain.RequestRecord, appEvents []domain.AppEvent, assertions []domain.ScenarioAssertion) []domain.ScenarioCheck {
	checks := make([]domain.ScenarioCheck, 0, len(assertions))
	var previousRequest *domain.RequestExpectation
	for _, assertion := range assertions {
		if assertion.AppEvent != nil {
			previousRequest = nil
			checks = append(checks, evaluateAppEventAssertion(appEvents, *assertion.AppEvent))
			continue
		}
		if assertion.Request != nil {
			previousRequest = assertion.Request
			name := assertion.Request.Method + " " + assertion.Request.Path + " observed"
			matched := false
			for _, record := range records {
				if strings.EqualFold(record.Method, assertion.Request.Method) && assertionPathMatches(assertion.Request.Path, record.Path) {
					matched = true
					break
				}
			}
			if matched {
				checks = append(checks, passedCheck(name))
			} else {
				checks = append(checks, domain.ScenarioCheck{Name: name, Message: "matching request was not observed"})
			}
			continue
		}
		name := fmt.Sprintf("HTTP %d observed", assertion.Response.Status)
		if previousRequest != nil {
			name = fmt.Sprintf("%s %s returned HTTP %d", previousRequest.Method, previousRequest.Path, assertion.Response.Status)
		}
		matched := false
		for _, record := range records {
			requestMatches := previousRequest == nil || (strings.EqualFold(record.Method, previousRequest.Method) && assertionPathMatches(previousRequest.Path, record.Path))
			if requestMatches && record.Status == assertion.Response.Status {
				matched = true
				break
			}
		}
		if matched {
			checks = append(checks, passedCheck(name))
		} else {
			message := "matching response was not observed"
			if previousRequest != nil {
				message = "the matching request was not observed with the expected response status"
			}
			checks = append(checks, domain.ScenarioCheck{Name: name, Message: message})
		}
	}
	return checks
}

func evaluateAppEventAssertion(events []domain.AppEvent, expectation domain.AppEventExpectation) domain.ScenarioCheck {
	name := fmt.Sprintf("app %s %s observed", expectation.Kind, expectation.Name)
	for _, event := range events {
		if event.Kind != expectation.Kind || event.Name != expectation.Name {
			continue
		}
		if expectation.Framework != "" && event.Framework != expectation.Framework {
			continue
		}
		if expectation.Kind == domain.AppEventAssertion {
			wantPassed := true
			if expectation.Passed != nil {
				wantPassed = *expectation.Passed
			}
			if event.Passed == nil || *event.Passed != wantPassed {
				return domain.ScenarioCheck{Name: name, Message: fmt.Sprintf("app assertion reported passed=%v", event.Passed != nil && *event.Passed)}
			}
		}
		return passedCheck(name)
	}
	return domain.ScenarioCheck{Name: name, Message: "matching app event was not observed"}
}

func hasAppEventAssertions(assertions []domain.ScenarioAssertion) bool {
	for _, assertion := range assertions {
		if assertion.AppEvent != nil {
			return true
		}
	}
	return false
}

func assertionPathMatches(pattern, observed string) bool {
	if pattern == observed {
		return true
	}
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")
	observedSegments := strings.Split(strings.Trim(observed, "/"), "/")
	if len(patternSegments) != len(observedSegments) {
		return false
	}
	for index, segment := range patternSegments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2 {
			if observedSegments[index] == "" {
				return false
			}
			continue
		}
		if segment != observedSegments[index] {
			return false
		}
	}
	return true
}

func recordsSince(records []domain.RequestRecord, started time.Time) []domain.RequestRecord {
	result := make([]domain.RequestRecord, 0, len(records))
	for _, record := range records {
		if !record.Timestamp.Before(started) {
			result = append(result, record)
		}
	}
	return result
}

func appEventsSince(events []domain.AppEvent, started time.Time) []domain.AppEvent {
	result := make([]domain.AppEvent, 0, len(events))
	for _, event := range events {
		if !event.Timestamp.Before(started) {
			result = append(result, event)
		}
	}
	return result
}

func passedCheck(name string) domain.ScenarioCheck {
	return domain.ScenarioCheck{Name: name, Passed: true}
}

func failedCheck(name string, err error) domain.ScenarioCheck {
	return domain.ScenarioCheck{Name: name, Passed: false, Message: err.Error()}
}

func allPassed(checks []domain.ScenarioCheck) bool {
	for _, check := range checks {
		if !check.Passed {
			return false
		}
	}
	return true
}
