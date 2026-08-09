package reporting

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func WriteScenarioTerminal(writer io.Writer, result domain.ScenarioResult) {
	fmt.Fprintf(writer, "Scenario: %s\n\n", result.Name)
	for _, check := range append(append([]domain.ScenarioCheck{}, result.Steps...), result.Assertions...) {
		mark := "✓"
		if !check.Passed {
			mark = "✗"
		}
		fmt.Fprintf(writer, "%s %s", mark, check.Name)
		if check.Message != "" {
			fmt.Fprintf(writer, ": %s", check.Message)
		}
		fmt.Fprintln(writer)
	}
	status := "PASSED"
	if !result.Passed {
		status = "FAILED"
	}
	fmt.Fprintf(writer, "\n%s\nDuration: %dms\n", status, result.DurationMS)
}

func WriteScenarioJSON(writer io.Writer, result domain.ScenarioResult) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
