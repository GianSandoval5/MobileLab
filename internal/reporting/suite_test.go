package reporting

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func testSuite() domain.ScenarioSuiteResult {
	started := time.Date(2026, 8, 9, 20, 0, 0, 0, time.UTC)
	return domain.NewScenarioSuiteResult("CI <suite>", started, 1250, []domain.ScenarioResult{
		{Name: "Passing scenario", Passed: true, StartedAt: started, DurationMS: 250, Steps: []domain.ScenarioCheck{{Name: "launch", Passed: true}}},
		{Name: "Failure <script>", StartedAt: started, DurationMS: 1000, Assertions: []domain.ScenarioCheck{{Name: "GET /profile", Message: "wanted 200 & received 500"}}},
	})
}

func TestJUnitReporterProducesConsumableEscapedXML(t *testing.T) {
	reporter, err := NewSuiteReporter(FormatJUnit)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := reporter.Write(&output, testSuite()); err != nil {
		t.Fatal(err)
	}
	var document junitSuites
	if err := xml.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("invalid XML:\n%s\n%v", output.String(), err)
	}
	if document.Tests != 2 || document.Failures != 1 || len(document.Suites) != 2 {
		t.Fatalf("unexpected JUnit totals: %#v", document)
	}
	if strings.Contains(output.String(), "<script>") || !strings.Contains(output.String(), "&lt;script&gt;") {
		t.Fatalf("JUnit content was not escaped: %s", output.String())
	}
}

func TestHTMLReporterIsStandaloneAndEscapesUserContent(t *testing.T) {
	reporter, err := NewSuiteReporter(FormatHTML)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := reporter.Write(&output, testSuite()); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	if !strings.HasPrefix(got, "<!doctype html>") || !strings.Contains(got, "2</strong>") || !strings.Contains(got, "1</strong>") {
		t.Fatalf("incomplete HTML report: %s", got)
	}
	if strings.Contains(got, "<script>") || !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("HTML content was not escaped: %s", got)
	}
}

func TestJSONSuiteReporterIncludesAggregateSummary(t *testing.T) {
	reporter, _ := NewSuiteReporter(FormatJSON)
	var output bytes.Buffer
	if err := reporter.Write(&output, testSuite()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"failed": 1`) || !strings.Contains(output.String(), `"scenarios"`) {
		t.Fatalf("missing suite summary: %s", output.String())
	}
}
