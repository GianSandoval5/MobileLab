package scenario

import (
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestYAMLParserCreatesPlatformNeutralDomain(t *testing.T) {
	input := []byte(`name: Payment with expired session
backend:
  latency: 200
  error: 500
auth:
  token: expired
device:
  network: online
steps:
  - launch_app
  - open_deeplink: myapp://payments
expect:
  - request:
      method: post
      path: /payments
  - response:
      status: 500
`)
	definition, err := (YAMLParser{}).Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if definition.Backend.LatencyMS != 200 || definition.Steps[1].Kind != domain.StepOpenDeepLink || definition.Assertions[0].Request.Method != "POST" {
		t.Fatalf("unexpected definition: %#v", definition)
	}
}

func TestYAMLParserRejectsUnknownAndUnsupportedOperations(t *testing.T) {
	for _, input := range []string{
		"name: test\nmystery: true\n",
		"name: test\nsteps: [teleport]\n",
		"name: test\nexpect:\n  - response: {status: 900}\n",
	} {
		if _, err := (YAMLParser{}).Parse([]byte(input)); err == nil {
			t.Errorf("expected parse failure for %q", input)
		}
	}
}

func TestYAMLParserRequiresDeepLinkValue(t *testing.T) {
	_, err := (YAMLParser{}).Parse([]byte("name: test\nsteps:\n  - open_deeplink: ''\n"))
	if err == nil || !strings.Contains(err.Error(), "requires a value") {
		t.Fatalf("unexpected error: %v", err)
	}
}
