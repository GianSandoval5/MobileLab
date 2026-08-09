package reporting

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
	Body    string `xml:",chardata"`
}

type junitCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitScenarioSuite struct {
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Skipped   int         `xml:"skipped,attr"`
	Time      string      `xml:"time,attr"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Cases     []junitCase `xml:"testcase"`
	SystemErr string      `xml:"system-err,omitempty"`
}

type junitSuites struct {
	XMLName  xml.Name             `xml:"testsuites"`
	Name     string               `xml:"name,attr"`
	Tests    int                  `xml:"tests,attr"`
	Failures int                  `xml:"failures,attr"`
	Time     string               `xml:"time,attr"`
	Suites   []junitScenarioSuite `xml:"testsuite"`
}

type junitSuiteReporter struct{}

func (junitSuiteReporter) Write(writer io.Writer, suite domain.ScenarioSuiteResult) error {
	document := junitSuites{Name: suite.Name, Time: seconds(suite.DurationMS)}
	for _, result := range suite.Scenarios {
		converted := junitScenarioSuite{
			Name:      result.Name,
			Time:      seconds(result.DurationMS),
			Timestamp: result.StartedAt.UTC().Format(time.RFC3339Nano),
			SystemErr: result.Error,
		}
		checks := append(append([]domain.ScenarioCheck(nil), result.Steps...), result.Assertions...)
		for _, check := range checks {
			testCase := junitCase{Name: check.Name, ClassName: result.Name}
			if !check.Passed {
				testCase.Failure = &junitFailure{Message: failureMessage(check.Message), Type: "assertion", Body: check.Message}
				converted.Failures++
			}
			converted.Cases = append(converted.Cases, testCase)
		}
		if len(converted.Cases) == 0 || (!result.Passed && converted.Failures == 0) {
			testCase := junitCase{Name: "scenario execution", ClassName: result.Name}
			if !result.Passed || result.Error != "" {
				testCase.Failure = &junitFailure{Message: failureMessage(result.Error), Type: "execution", Body: result.Error}
				converted.Failures++
			}
			converted.Cases = append(converted.Cases, testCase)
		}
		converted.Tests = len(converted.Cases)
		document.Tests += converted.Tests
		document.Failures += converted.Failures
		document.Suites = append(document.Suites, converted)
	}
	if _, err := io.WriteString(writer, xml.Header); err != nil {
		return err
	}
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(document); err != nil {
		return fmt.Errorf("encode JUnit XML: %w", err)
	}
	if _, err := io.WriteString(writer, "\n"); err != nil {
		return err
	}
	return nil
}

func seconds(milliseconds int64) string {
	return strconv.FormatFloat(float64(milliseconds)/1000, 'f', 3, 64)
}

func failureMessage(message string) string {
	if message == "" {
		return "MobileLab scenario failed"
	}
	return message
}
