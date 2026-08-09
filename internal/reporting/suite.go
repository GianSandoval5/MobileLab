package reporting

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type Format string

const (
	FormatTerminal Format = "terminal"
	FormatJSON     Format = "json"
	FormatJUnit    Format = "junit"
	FormatHTML     Format = "html"
)

func (format Format) Valid() bool {
	switch format {
	case FormatTerminal, FormatJSON, FormatJUnit, FormatHTML:
		return true
	default:
		return false
	}
}

type SuiteReporter interface {
	Write(io.Writer, domain.ScenarioSuiteResult) error
}

func NewSuiteReporter(format Format) (SuiteReporter, error) {
	switch format {
	case FormatTerminal:
		return terminalSuiteReporter{}, nil
	case FormatJSON:
		return jsonSuiteReporter{}, nil
	case FormatJUnit:
		return junitSuiteReporter{}, nil
	case FormatHTML:
		return htmlSuiteReporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported report format %q", format)
	}
}

type terminalSuiteReporter struct{}

func (terminalSuiteReporter) Write(writer io.Writer, suite domain.ScenarioSuiteResult) error {
	for index, result := range suite.Scenarios {
		if index > 0 {
			fmt.Fprintln(writer)
		}
		WriteScenarioTerminal(writer, result)
	}
	if len(suite.Scenarios) > 1 {
		fmt.Fprintf(writer, "\nSuite: %s\nScenarios: %d passed, %d failed, %d total\nDuration: %dms\n", suite.Name, suite.Summary.Passed, suite.Summary.Failed, suite.Summary.Total, suite.DurationMS)
	}
	return nil
}

type jsonSuiteReporter struct{}

func (jsonSuiteReporter) Write(writer io.Writer, suite domain.ScenarioSuiteResult) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(suite)
}
